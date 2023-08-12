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
	"github.com/samber/mo"
	"xorm.io/xorm"
)

type GithubSearchResponse struct {
	TotalCount         int64     `json:"total_count"`
	Items              []Repo    `json:"items"`
	IncompleteResults  bool      `json:"incomplete_results"`
	RatelimitRemaining int       `json:"-"`
	RatelimitReset     time.Time `json:"-"`
	Page               int       `json:"-"`
	NotModified        bool      `json:"-"`
}

type GithubSearchResponseError struct {
	Error error
	Page  int
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

func minStarsQuery(minStars int64) string {
	return fmt.Sprintf("stars:>%v", minStars)
}

func minMaxStarsQuery(minStars int64, maxStars int64) string {
	if minStars == maxStars {
		return fmt.Sprintf("stars:%v", minStars)
	} else {
		return fmt.Sprintf("stars:%v..%v", minStars, maxStars)
	}
}

func createdOnQuery(creation RepoCreationDateRange) string {
	if creation.CoversEverything() {
		return ""
	} else {
		return "created:" + creation.ToQueryString()
	}
}

func search(githubClient *http.Client, page int, searchTerm ...string) GithubSearchResponse {
	query := strings.Trim(strings.Join(searchTerm, " "), " ")

	log.Printf("[search] searching with terms '%s' - page %d\n", query, page)

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
		log.Panicf("Received response code %v from github. Response body: %v", response.Status, string(lo.Must(io.ReadAll(response.Body))))
	}

	if *enableResponsesLog {
		resLogger.Println(string(lo.Must(httputil.DumpResponse(response, false))))
	}

	body := lo.Must(io.ReadAll(response.Body))

	decodedResponse := GithubSearchResponse{}
	lo.Must0(json.Unmarshal(body, &decodedResponse), "Could not decode response json")
	decodedResponse.Page = page
	decodedResponse.RatelimitRemaining = lo.Must(strconv.Atoi(response.Header.Get("X-Ratelimit-Remaining")))
	decodedResponse.RatelimitReset = time.Unix(lo.Must(strconv.ParseInt(response.Header.Get("X-Ratelimit-Reset"), 10, 64)), 0)

	return decodedResponse
}

func searchToChannel(client *http.Client, maybeResponse chan<- mo.Either[GithubSearchResponse, GithubSearchResponseError], page int, searchTerm ...string) {
	err, ok := lo.TryWithErrorValue(func() error {
		maybeResponse <- mo.Left[GithubSearchResponse, GithubSearchResponseError](search(client, page, searchTerm...))
		log.Printf("[async] Got response for page %v\n", page)
		return nil
	})
	if !ok {
		maybeResponse <- mo.Right[GithubSearchResponse, GithubSearchResponseError](GithubSearchResponseError{
			Error: errors.Errorf("Could not fetch async search: %v", err),
			Page:  page,
		})
	}
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

	firstPage := search(client, 1, minMaxStarsQuery(minStars, maxStars), createdOnQuery(creationDateRange))
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
			log.Println("[zero] No results are present, the date range covers today - decreasing maxStars by 1 and increasing the window size")
			SetMaxStars(db, maxStars-1)
			increaseWindowSize(db)
			return
		} else {
			log.Println("[zero] No results are present, the date range doesn't cover today - going to the next date range and increasing date range size")
			creationDateRange.NextRange().BiggerRange().Save(db)
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
	} else if firstPage.IncompleteResults && firstPage.TotalCount >= MAX_RESULTS_PER_PAGE*(MAX_PAGES/2) && searchWindow != 0 {
		// even if we aren't really close to the 10-page limit but IncompleteResults is set, let's try to make IncompleteResults go away
		decreaseWindowSize(db)
	} else if firstPage.TotalCount <= MAX_RESULTS_PER_PAGE*4 && creationDateRange.CoversEverything() {
		// only up to four pages - can definitely increase the window size now
		increaseWindowSize(db)
	} else if firstPage.TotalCount <= MAX_RESULTS_PER_PAGE*4 && !creationDateRange.CoversToday() {
		// the same - but when filtering on dates (and this isn't the last page)
		creationDateRange.BiggerRange().Save(db)
	}

	// we already have the first page (have to get it first synchronously to get the number of pages), so start from the second one
	startAtPage := 2
	startFetchAtPage := startAtPage
	pagesLeftToProcess := pages - startAtPage + 1

	maybeResponses := make(chan mo.Either[GithubSearchResponse, GithubSearchResponseError], pagesLeftToProcess)

	// If we don't have enough Ratelimit - let's first do what we can asynchronously, waiting when needed
	if firstPage.RatelimitRemaining <= pagesLeftToProcess {
		firstBatchSize := min(firstPage.RatelimitRemaining, pagesLeftToProcess)
		maybeResponsesBeforeRatelimit := make(chan mo.Either[GithubSearchResponse, GithubSearchResponseError], firstBatchSize)
		for i := 0; i < firstBatchSize; i++ {
			go searchToChannel(client, maybeResponsesBeforeRatelimit, startFetchAtPage+i, minMaxStarsQuery(minStars, maxStars), createdOnQuery(creationDateRange))
		}
		for i := 0; i < firstBatchSize; i++ {
			maybeResponse, ok := <-maybeResponsesBeforeRatelimit
			lo.Must0(ok, "[async] Did not receive enough objects in the maybeResponsesBeforeRatelimit channel")
			maybeResponse.ForEach(
				func(resp GithubSearchResponse) { lo.Must0(resp.WaitIfNeccessary(ctx)) },
				func(_ GithubSearchResponseError) { log.Println("[async] passing the error up") })
			maybeResponses <- maybeResponse
		}
		startFetchAtPage += firstBatchSize
	}

	// And now either fetch all the pages or (if we were low on Ratelimit) fetch the pages left
	for i := startFetchAtPage; i <= pages; i++ {
		go searchToChannel(client, maybeResponses, i, minMaxStarsQuery(minStars, maxStars), createdOnQuery(creationDateRange))
	}

	savedMaybeResponses := make([]mo.Result[GithubSearchResponse], pagesLeftToProcess)

	for i := 0; i < pagesLeftToProcess; i++ {
		maybeResponse, ok := <-maybeResponses
		lo.Must0(ok, "[async] Did not receive enough maybeResponses")
		maybeResponse.ForEach(
			func(response GithubSearchResponse) {
				log.Printf("[async] Processing response for page %d\n", response.Page)
				save(db, response)
				savedMaybeResponses[response.Page-startAtPage] = mo.Ok(response)
			}, func(error GithubSearchResponseError) {
				log.Printf("[async] Skipping saving for page %d because of an error\n", error.Page)
				savedMaybeResponses[error.Page-startAtPage] = mo.Err[GithubSearchResponse](error.Error)
			})
	}

	for _, savedMaybeResponse := range savedMaybeResponses {
		// Only now crash on an error - after every result is saved and max stars and date ranges are decreased
		// for all the previous ones
		response := savedMaybeResponse.MustGet()
		if !response.IncompleteResults && response.Page == pages {
			if creationDateRange.CoversToday() {
				// this is the last page - we are sure nothing was missed, can decrease to one beyond minimum
				decreaseMaxStarsBeyondMinimum(db, response)
			} else {
				creationDateRange.NextRange().Save(db)
			}
		} else {
			decreaseMaxStarsToMinumum(db, response)
		}
	}
}

func save(db *xorm.Engine, resp GithubSearchResponse) {
	lo.Must(db.Transaction(func(tx *xorm.Session) (any, error) {
		for _, repo := range resp.Items {
			repo.LastFetchedFromGithubAt = time.Now()
			repo.NotSeenSinceCounter = 0
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
	log.Printf("Fetching and saving the top of the top - repositiories with at least %d stars\n", MAX_STARS_DEFAULT)
	minStars := int64(MAX_STARS_DEFAULT)

	resp := search(client, 1, minStarsQuery(minStars))
	lo.Must0(resp.WaitIfNeccessary(ctx))
	save(db, resp)

	pages := numberOfPages(resp.TotalCount)

	if pages > MAX_PAGES {
		panic("There's too many 'top of the top' repositories - increase the limit!")
	}

	for page := 2; page <= pages; page += 1 {
		res := search(client, page, minStarsQuery(minStars))
		lo.Must0(res.WaitIfNeccessary(ctx))
		save(db, res)
	}
}
