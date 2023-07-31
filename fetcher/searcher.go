package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/samber/lo/parallel"
	"github.com/samber/mo"
	"xorm.io/xorm"
)

type GithubSearchResponse struct {
	TotalCount         int64     `json:"total_count"`
	Items              []Repo    `json:"items"`
	IncompleteResults  bool      `json:"incomplete_results"`
	RatelimitRemaining int       `json:"-"`
	RatelimitReset     time.Time `json:"-"`
}

// The builtin time.Sleep function doesn't take context.Context into account
func sleepContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func (resp *GithubSearchResponse) WaitIfNeccessary(ctx context.Context) error {
	if resp.RatelimitRemaining <= 0 {
		until := time.Until(resp.RatelimitReset)

		// For safety - assume a bit of clock drift
		until += 4 * time.Second

		if until.Microseconds() >= 0 {
			log.Printf("Sleeping for %v seconds", until)
			return sleepContext(ctx, until)
		} else {
			duration := 20 * time.Second
			log.Printf("Wanted to sleep for negative time: %v. Sleeping for %v seconds instead", until, duration)
			return sleepContext(ctx, duration)
		}
	}

	return nil
}

func search(githubClient *http.Client, minStars, maxStars int64, page int, searchTerm ...string) GithubSearchResponse {
	if minStars == maxStars {
		log.Printf("[search] searching for %v stars, page: %v, extra search terms: %v\n",
			minStars, page, searchTerm)
	} else {
		log.Printf("[search] searching from %v to %v stars, page: %v, extra search terms: %v\n",
			minStars, maxStars, page, searchTerm)
	}

	query := fmt.Sprintf("stars:%v..%v", minStars, maxStars)
	if len(searchTerm) != 0 {
		query += " " + strings.Join(searchTerm, " ")
	}

	reqUrl := lo.Must(url.Parse("https://api.github.com/search/repositories"))

	reqUrlParams := reqUrl.Query()
	reqUrlParams.Set("q", query)
	reqUrlParams.Set("sort", "stars")
	reqUrlParams.Set("per_page", strconv.Itoa(MAX_RESULTS_PER_PAGE))
	reqUrlParams.Set("page", strconv.Itoa(page))
	reqUrl.RawQuery = reqUrlParams.Encode()

	req := lo.Must(http.NewRequest("GET", reqUrl.String(), nil))
	if *enableRequestLog {
		reqLogger.Println(string(lo.Must(httputil.DumpRequest(req, false))))
	}

	response := lo.Must(githubClient.Do(req))
	if response.StatusCode != http.StatusOK {
		panic("Received response code " + response.Status + " from github")
	}

	if *enableResponsesLog {
		resLogger.Println(string(lo.Must(httputil.DumpResponse(response, false))))
	}

	body := lo.Must(io.ReadAll(response.Body))

	decodedResponse := GithubSearchResponse{}
	lo.Must0(json.Unmarshal(body, &decodedResponse), "Could not decode response json")
	decodedResponse.RatelimitRemaining = lo.Must(strconv.Atoi(response.Header.Get("X-Ratelimit-Remaining")))
	decodedResponse.RatelimitReset = time.Unix(lo.Must(strconv.ParseInt(response.Header.Get("X-Ratelimit-Reset"), 10, 64)), 0)

	return decodedResponse
}

func searchWithCreationDate(ghClient *http.Client, minStars, maxStars int64, creation RepoCreationDateRange, page int) GithubSearchResponse {
	return search(ghClient, minStars, maxStars, page, "created:"+creation.ToQueryString())
}

func smallerWindow(window int64) int64 {
	if window == 0 {
		log.Panic("Trying to decrease window size from 0")
	}
	if window <= 2 {
		return window - 1
	} else {
		return int64(math.Round(float64(window) * 0.5))
	}
}

func biggerWindow(window int64) int64 {
	if window <= 2 {
		return window + 1
	}
	return int64(math.Round(float64(window) * 1.5))
}

func decreaseWindowSize(db *xorm.Engine) {

	oldSearchWindow := GetSearchWindow(db)
	newSearchWindow := smallerWindow(oldSearchWindow)
	log.Printf("Decreasing window size from %v to %v", oldSearchWindow, newSearchWindow)

	SetSearchWindow(db, newSearchWindow)
}

func increaseWindowSize(db *xorm.Engine) {
	oldSearchWindow := GetSearchWindow(db)
	newSearchWindow := biggerWindow(oldSearchWindow)

	log.Printf("Increasing window size from %v to %v", oldSearchWindow, newSearchWindow)

	SetSearchWindow(db, newSearchWindow)
}

func doFetcherTask(ctx context.Context, client *http.Client, db *xorm.Engine) {
	maxStars, searchWindow := GetMaxStars(db), GetSearchWindow(db)
	minStars := maxStars - searchWindow
	creationDateRange := GetRepoCreationDateRange(db)

	log.Println("-- Fetcher --")
	log.Printf("-- Stars: [from %v to %v] -> window=%v, creation date range: %v --\n", minStars, maxStars, searchWindow, creationDateRange.ToQueryString())

	if creationDateRange.howManyDays < 0 {
		log.Printf("creationDateRange.howManyDays got set to %v, resetting to 1\n", creationDateRange.howManyDays)
		creationDateRange.howManyDays = 1
		creationDateRange.Save(db)
		return
	}

	if searchWindow < 0 {
		log.Printf("Search window got set to %v, resetting to 1\n", searchWindow)
		SetSearchWindow(db, 1)
		return
	}

	if searchWindow >= maxStars {
		log.Printf("Search window got bigger than maxStars (%v > %v) - capping to %v\n", searchWindow, maxStars, maxStars-1)
		SetSearchWindow(db, maxStars-1)
		return
	}

	if maxStars >= MAX_STARS_DEFAULT {
		log.Println("switching from fetching very big repos to big ones")
		fetchAndSaveReposWithVeryHighStarsCount(db, client, ctx)
		SetMaxStars(db, maxStars-1)
		return
	}

	// TODO: handle IncompleteResults == true

	firstPage := searchWithCreationDate(client, minStars, maxStars, creationDateRange, 1)
	lo.Must0(firstPage.WaitIfNeccessary(ctx))
	save(db, firstPage)

	log.Printf("Got %v results\n", firstPage.TotalCount)
	if firstPage.IncompleteResults {
		log.Println("Warning, incomplete_results is set! TODO!")
	}

	if firstPage.TotalCount > MAX_RESULTS_PER_PAGE*MAX_PAGES {
		if searchWindow > 0 {
			// we might be missing some results, redo the same search later with a decreased
			// window size to get them
			decreaseWindowSize(db)
		} else if creationDateRange.howManyDays > 0 {
			// Search star window is 0, cannot decrease it anymore.
			// Start decreasing the creation days window
			creationDateRange.HalvedRange().Save(db)
		} else {
			panic("Cannot make the query any more specific!")
		}
		// Don't request other result pages - something might be missing
		return
	}

	pages := numberOfPages(firstPage.TotalCount)

	// Only do after checking if TotalCount wasn't overflowed
	if pages == 0 {
		if creationDateRange.CoversToday() {
			log.Println("[zero] No results are present, the date range covers today - decreasing maxStars by 1")
			SetMaxStars(db, maxStars-1)
			return
		} else {
			log.Println("[zero] No results are present, the date range doesn't cover today - going to the next date range")
			creationDateRange.NextRange().Save(db)
			return
		}
	}

	if !firstPage.IncompleteResults && pages == 1 {
		// this is the only page - we are sure nothing was missed, can decrease to one beyond minimum
		if creationDateRange.CoversToday() {
			decreaseMaxStarsBeyondMinimum(db, firstPage)
		} else {
			creationDateRange.NextRange().Save(db)
		}
	} else {
		// There are still results left to fetch for this amount of stars
		decreaseMaxStarsToMinumum(db, firstPage)
	}

	if firstPage.TotalCount > MAX_RESULTS_PER_PAGE*(MAX_PAGES-2) && searchWindow != 0 {
		// we got pretty close to the limit - but no repositories should be missing due to
		// the result fitting in the 1000 responses limit
		decreaseWindowSize(db)
	} else if firstPage.TotalCount <= MAX_RESULTS_PER_PAGE*4 && creationDateRange.CoversEverything() {
		// only up to four pages - can definitely increase the window size now
		increaseWindowSize(db)
	} else if firstPage.TotalCount <= MAX_RESULTS_PER_PAGE*4 && !creationDateRange.CoversToday() {
		// the same - but when filtering on dates (and this isn't the last page)
		creationDateRange.BiggerRange().Save(db)
	}

	startAtPage := 2
	// If we have the Ratelimit resources to do so - search everything but the first and last page in parallel
	if pages >= 4 && firstPage.RatelimitRemaining > pages-1 {
		maybeResponses := parallel.Times(pages-2, func(index int) mo.Result[GithubSearchResponse] {
			var response GithubSearchResponse
			err, ok := lo.TryWithErrorValue(func() error {
				response = searchWithCreationDate(newGithubApiClient(ctx), minStars, maxStars, creationDateRange, index+2)
				log.Printf("[async] Got response for page %v\n", index+2)
				return nil
			})
			if !ok {
				return mo.Err[GithubSearchResponse](errors.Errorf("Could not fetch async search: %v", err))
			} else {
				return mo.Ok(response)
			}
		})

		for i, maybeResponse := range maybeResponses {
			response := maybeResponse.MustGet()
			log.Printf("[async] Processing response %v after async loop\n", i)
			lo.Must0(response.WaitIfNeccessary(ctx))
			save(db, response)
			decreaseMaxStarsToMinumum(db, response)
		}

		startAtPage = pages // the last page
	}

	// And now either fetch the last page or (if low on Ratelimit) fetch pages synchronously to properly wait
	for page := startAtPage; page <= pages; page += 1 {
		res := searchWithCreationDate(client, minStars, maxStars, creationDateRange, page)
		lo.Must0(res.WaitIfNeccessary(ctx))
		save(db, res)
		if !res.IncompleteResults && page == pages {
			if creationDateRange.CoversToday() {
				// this is the last page - we are sure nothing was missed, can decrease to one beyond minimum
				decreaseMaxStarsBeyondMinimum(db, res)
			} else {
				creationDateRange.NextRange().Save(db)
			}
		} else {
			decreaseMaxStarsToMinumum(db, res)
		}
	}
}

func save(db *xorm.Engine, resp GithubSearchResponse) {
	lo.Must(db.Transaction(func(tx *xorm.Session) (any, error) {
		for _, repo := range resp.Items {
			repo.LastFetchedFromGithubAt = time.Now()
			if lo.Must(tx.Exist(&Repo{Id: repo.Id})) {
				lo.Must(tx.ID(repo.Id).AllCols().Update(repo))
			} else {
				repo.FirstFetchedFromGithubAt = time.Now()
				lo.Must(tx.Insert(repo))
			}

			// Get rid of repositories with a different ID than just inserted, but with the same FullName
			// This happens when a repository is deleted, but a new one with the same name is created in its place
			affected := lo.Must(tx.Where("Id != ? and FullName = ?", repo.Id, repo.FullName).Unscoped().Delete(&Repo{}))
			if affected > 0 {
				log.Printf("[save] Deleted %d duplicate entries for repo %s", affected, repo.FullName)
			}
		}
		return nil, nil
	}))
}

func decreaseMaxStarsToMinumum(db *xorm.Engine, resp GithubSearchResponse) {
	lo.Must(db.Transaction(func(tx *xorm.Session) (any, error) {
		if len(resp.Items) > 0 {
			leastStargazers := resp.Items[len(resp.Items)-1].Stargazers
			maxStars := GetMaxStars(tx)
			if leastStargazers > maxStars {
				log.Printf("[!!!] [Warning] Trying to <increase> maxStars in decreaseMaxStars (from %v to %v)"+
					" - GitHub's api might be a bit weird\n", maxStars, leastStargazers)
			}
			if leastStargazers < maxStars {
				log.Printf("Decreasing max stars from %v to %v\n", GetMaxStars(db), leastStargazers)
				SetMaxStars(tx, leastStargazers)
			}
		}
		return nil, nil
	}))
}

func decreaseMaxStarsBeyondMinimum(db *xorm.Engine, resp GithubSearchResponse) {
	lo.Must(db.Transaction(func(tx *xorm.Session) (any, error) {
		if len(resp.Items) > 0 {
			leastStargazers := resp.Items[len(resp.Items)-1].Stargazers
			setTo := leastStargazers - 1
			if setTo != GetMaxStars(tx) {
				log.Printf("Decreasing max stars from %v to %v (going 1 beyond last stargazers - this is the last page)\n", GetMaxStars(db), setTo)
				SetMaxStars(tx, setTo)
			}
		}
		return nil, nil
	}))
}

func numberOfPages(results int64) int {
	pages := results / 100
	if results%100 != 0 {
		pages += 1
	}
	return int(pages)
}

func fetchAndSaveReposWithVeryHighStarsCount(db *xorm.Engine, client *http.Client, ctx context.Context) {
	log.Println("Fetching and saving the top of the top - repositiories with at least 50 thousand stars")
	minStars, maxStars := int64(MAX_STARS_DEFAULT), int64(10000000)

	resp := search(client, minStars, maxStars, 1)
	lo.Must0(resp.WaitIfNeccessary(ctx))
	save(db, resp)

	pages := numberOfPages(resp.TotalCount)

	if pages > MAX_PAGES {
		panic("There's too many 'top of the top' repositories - increase the limit!")
	}

	for page := 2; page <= pages; page += 1 {
		res := search(client, minStars, maxStars, page)
		lo.Must0(res.WaitIfNeccessary(ctx))
		save(db, res)
	}
}
