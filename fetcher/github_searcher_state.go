package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/samber/lo"
	"xorm.io/xorm"
)

const (
	MAX_STARS_KEY               = "max_stars"
	MAX_STARS_DEFAULT           = 200000
	SEARCH_WINDOW_KEY           = "search_window"
	SEARCH_WINDOW_DEFAULT       = 10000
	DATE_START_SECOND_KEY       = "second_start"
	DATE_SECONDS_WINDOW_KEY     = "seconds_window"
	GETREPO_RATELIMIT_RESET     = "getrepo_ratelimit_reset"
	GETREPO_RATELIMIT_REMAINING = "getrepo_ratelimit_remaining"
	DEFAULT_GETREPO_LIMIT       = 6000
)

func getFromState[T any](db xorm.Interface, key string, defaultValue T) T {
	state := &State{Name: key}
	if lo.Must(db.Get(state)) {
		value := new(T)
		lo.Must0(json.Unmarshal([]byte(state.Value), value))
		return *value
	} else {
		return defaultValue
	}
}

func setToState[T any](db xorm.Interface, key string, value T) {
	lo.Must(db.Exec("INSERT OR REPLACE INTO State(Name, Value) VALUES(?, ?)", key, string(lo.Must(json.Marshal(value)))))
}

func GetMaxStars(db xorm.Interface) int64 {
	return getFromState[int64](db, MAX_STARS_KEY, MAX_STARS_DEFAULT)
}
func SetMaxStars(db xorm.Interface, stars int64) {
	// TODO: this check could probably be done in a transaction, however only one process can currently be a fetcher, so whatever
	if stars != GetMaxStars(db) {
		log.Println("Changed max stars - resetting SearchDaysWindow and SearchStartDay")
		DefaultRepoCreationDateRange().Save(db)
	}
	setToState[int64](db, MAX_STARS_KEY, stars)
}

func GetSearchWindow(db xorm.Interface) int64 {
	return getFromState[int64](db, SEARCH_WINDOW_KEY, SEARCH_WINDOW_DEFAULT)
}
func SetSearchWindow(db xorm.Interface, win int64) {
	setToState[int64](db, SEARCH_WINDOW_KEY, win)
}

func SetRepoRatelimit(db xorm.Interface, ratelimitReset time.Time, ratelimitRemaining int) {
	setToState[int64](db, GETREPO_RATELIMIT_RESET, ratelimitReset.Unix())
	setToState[int](db, GETREPO_RATELIMIT_REMAINING, ratelimitRemaining)
}
func GetRepoRatelimit(db xorm.Interface) (ratelimitReset time.Time, ratelimitRemaining int) {
	defaultReset := time.Now().Add(1 * time.Hour).Unix()
	ratelimitReset = time.Unix(getFromState[int64](db, GETREPO_RATELIMIT_RESET, defaultReset), 0)
	ratelimitRemaining = getFromState[int](db, GETREPO_RATELIMIT_REMAINING, DEFAULT_GETREPO_LIMIT)
	return
}

type RepoCreationDateRange struct {
	startingSecond int64
	howManySeconds int64
}

func githubCreationDay() time.Time {
	return time.Date(2007, time.November, 1, 12, 0, 0, 0, time.UTC)
}

func DefaultRepoCreationDateRange() RepoCreationDateRange {
	aboutToday := time.Now().AddDate(0, 0, 2)
	githubExistenceSeconds := int64(aboutToday.Sub(githubCreationDay()).Seconds())
	return RepoCreationDateRange{
		startingSecond: 0,
		howManySeconds: githubExistenceSeconds,
	}
}

func (r RepoCreationDateRange) CoversEverything() bool {
	return r.startingSecond == DefaultRepoCreationDateRange().startingSecond && r.CoversToday()
}

func GetRepoCreationDateRange(db *xorm.Engine) (r RepoCreationDateRange) {
	r.startingSecond = getFromState[int64](db, DATE_START_SECOND_KEY, -1)
	if r.startingSecond == -1 {
		return DefaultRepoCreationDateRange()
	}
	r.howManySeconds = getFromState[int64](db, DATE_SECONDS_WINDOW_KEY, -1)
	if r.howManySeconds == -1 {
		return DefaultRepoCreationDateRange()
	}
	return r
}

func (r RepoCreationDateRange) ToQueryString() string {
	start := githubCreationDay().Add(time.Duration(r.startingSecond) * time.Second)
	end := start.Add(time.Duration(r.howManySeconds) * time.Second)
	return fmt.Sprintf("%v..%v", start.Format("2006-01-02T15:04:05Z"), end.Format("2006-01-02T15:04:05Z"))
}

func (r RepoCreationDateRange) Save(db xorm.Interface) {
	// todo: transaction?
	setToState[int64](db, DATE_START_SECOND_KEY, r.startingSecond)
	setToState[int64](db, DATE_SECONDS_WINDOW_KEY, r.howManySeconds)
}

func (r RepoCreationDateRange) HalvedRange() (ret RepoCreationDateRange) {
	ret.startingSecond = r.startingSecond
	ret.howManySeconds = smallerWindow(r.howManySeconds)
	return ret
}

func (r RepoCreationDateRange) BiggerRange() (ret RepoCreationDateRange) {
	ret.startingSecond = r.startingSecond
	ret.howManySeconds = biggerWindow(r.howManySeconds)
	return ret
}

func (r RepoCreationDateRange) CoversToday() bool {
	end := githubCreationDay().
		Add(time.Duration(r.startingSecond) * time.Second).
		Add(time.Duration(r.howManySeconds) * time.Second)
	// add an hour just in case of either some one-of errors somewhere, or very new repos
	return end.After(time.Now().Add(1 * time.Hour))
}

func (r RepoCreationDateRange) NextRange() (ret RepoCreationDateRange) {
	ret.startingSecond = r.startingSecond + r.howManySeconds + 1
	ret.howManySeconds = r.howManySeconds
	log.Printf("[creationDateRange] Going from %v to %v\n", r, ret)
	return ret
}
