package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/samber/lo"
	"github.com/samber/lo/parallel"
	"xorm.io/xorm"
)

const HOW_MANY_REPOS_TO_CHECK_AT_ONCE_MAX = 25

func notModified(response *http.Response) bool {
	return response.StatusCode == http.StatusNotModified
}
func surelyExists(response *http.Response) bool {
	return response.StatusCode == http.StatusOK
}
func surelyDoesntExist(response *http.Response) bool {
	return response.StatusCode == http.StatusNotFound || // Deleted repositories
		response.StatusCode == http.StatusForbidden || // "Access to this repository has been disabled by GitHub Staff due to a violation of GitHub's terms of service"
		response.StatusCode == http.StatusUnavailableForLegalReasons // "This repository is currently disabled due to a DMCA takedown notice."
}

func getRepo(githubClient *http.Client, previouslyFetchedRepo Repo) GithubSearchResponse {
	result := GithubSearchResponse{
		IncompleteResults: true,
	}

	reqUrl := lo.Must(url.Parse(fmt.Sprintf("https://api.github.com/repos/%s", previouslyFetchedRepo.FullName)))
	req := lo.Must(http.NewRequest("GET", reqUrl.String(), nil))
	if previouslyFetchedRepo.GetRepoApiLastModifiedHeader != "" {
		req.Header.Set("If-Modified-Since", previouslyFetchedRepo.GetRepoApiLastModifiedHeader)
	}

	if *enableRequestLog {
		reqLogger.Println(string(lo.Must(httputil.DumpRequest(req, false))))
	}

	response, err := githubClient.Do(req)
	if err != nil {
		log.Println(fmt.Errorf("[deleter] getRepo: could not fetch from github: %w", err))
		return result
	}

	if *enableResponsesLog {
		resLogger.Println(string(lo.Must(httputil.DumpResponse(response, false))))
	}

	if !surelyExists(response) && !notModified(response) && !surelyDoesntExist(response) {
		log.Printf("[deleter] getRepo: Received response code %s from github\n", response.Status)
		return result
	}

	result.RatelimitRemaining, err = strconv.Atoi(response.Header.Get("X-Ratelimit-Remaining"))
	if err != nil {
		log.Println(fmt.Errorf("[deleter] getRepo: Could not convert X-Ratelimit-Remaining to int: %w", err))
		return result
	}

	ratelimitReset, err := strconv.ParseInt(response.Header.Get("X-Ratelimit-Reset"), 10, 64)
	if err != nil {
		log.Println(fmt.Errorf("[deleter] getRepo: Could not convert X-Ratelimit-Reset to int: %w", err))
		return result
	}
	result.RatelimitReset = time.Unix(ratelimitReset, 0)

	if surelyDoesntExist(response) {
		result.TotalCount = 0
		result.Items = []Repo{}
		result.IncompleteResults = false
	} else if notModified(response) {
		// If no changes: return the repository with the same old metadata, but just the last-modified header changed
		previouslyFetchedRepo.GetRepoApiLastModifiedHeader = response.Header.Get("Last-Modified")
		result.TotalCount = 1
		result.Items = []Repo{previouslyFetchedRepo}
		result.NotModified = true
		result.IncompleteResults = false
	} else if surelyExists(response) {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Println(fmt.Errorf("[deleter] getRepo: Could not read response body: %w", err))
			return result
		}

		repo := Repo{}
		err = json.Unmarshal(body, &repo)
		if err != nil {
			log.Println(fmt.Errorf("[deleter] getRepo: Could not decode repo json from GitHub: %w", err))
			return result
		}

		repo.GetRepoApiLastModifiedHeader = response.Header.Get("Last-Modified")

		// Note: the ID returned by the GitHub /repo endpoint is different from the one returned from the
		// /search endpoint. Let's override the ID to the search one here to be sure to never use the wrong
		// one anywhere
		repo.Id = previouslyFetchedRepo.Id

		result.TotalCount = 1
		result.Items = []Repo{repo}
		result.IncompleteResults = false
	} else {
		panic("Shouldn't get here: the check for failed request happens above.")
	}

	return result
}

func getLikelyDeletedRepositories(db *xorm.Engine, howMany int) (repos []Repo) {
	lo.Must0(db.
		Where("NotSeenSinceCounter > 2").
		Desc("NotSeenSinceCounter").
		Asc("Id").
		Limit(howMany).
		Find(&repos))

	return repos
}

func checkReposForDeletion(githubApiClient *http.Client, db *xorm.Engine) {
	previousRatelimitReset, previousRatelimitRemaining := GetRepoRatelimit(db)
	isPreviousRatelimitStillAccurate := time.Until(previousRatelimitReset) > -3*time.Second

	if previousRatelimitRemaining <= 0 && isPreviousRatelimitStillAccurate {
		log.Printf("[deleter] Skipping checking for deleted repositories: ran out of ratelimit, will not check in the next %d seconds\n",
			time.Until(previousRatelimitReset)/time.Second)
		return
	}

	var parallelFetches int
	if isPreviousRatelimitStillAccurate {
		parallelFetches = min(HOW_MANY_REPOS_TO_CHECK_AT_ONCE_MAX, previousRatelimitRemaining)
	} else {
		parallelFetches = min(HOW_MANY_REPOS_TO_CHECK_AT_ONCE_MAX, DEFAULT_GETREPO_LIMIT)
	}

	likelyDeleted := getLikelyDeletedRepositories(db, parallelFetches)
	log.Printf("[deleter] Checking %d likely-deleted repositories (asked the database for max %d)\n", len(likelyDeleted), parallelFetches)

	responses := parallel.Map(likelyDeleted, func(repo Repo, index int) GithubSearchResponse {
		return getRepo(githubApiClient, repo)
	})

	deletedCount, notModifiedCount, updatedCount := 0, 0, 0

	for i, response := range responses {
		if response.IncompleteResults {
			log.Printf("[deleter] Failed while fetching '%s' - skipping\n", likelyDeleted[i].FullName)
			continue
		}

		if response.TotalCount == 0 {
			lo.Must(db.ID(likelyDeleted[i].Id).Unscoped().Delete(&Repo{}))
			deletedCount++
		} else {
			save(db, response)

			if response.NotModified {
				notModifiedCount++
			} else {
				updatedCount++
			}
		}
	}

	if deletedCount > 0 || notModifiedCount > 0 || updatedCount > 0 {
		ratelimitReset := time.Unix(0, 0)
		ratelimitRemaining := 0

		for _, response := range responses {
			if response.IncompleteResults {
				continue
			}

			if response.RatelimitReset.Unix() == ratelimitReset.Unix() {
				ratelimitRemaining = min(ratelimitRemaining, response.RatelimitRemaining)
			} else if response.RatelimitReset.After(ratelimitReset) {
				ratelimitReset = response.RatelimitReset
				ratelimitRemaining = response.RatelimitRemaining
			}
		}

		SetRepoRatelimit(db, ratelimitReset, ratelimitRemaining)
	}

	log.Printf("[deleter] Done checking deleted repos: deleted %d, updated %d repositories, %d were not modified\n",
		deletedCount, updatedCount, notModifiedCount)
}
