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
	SEARCH_WINDOW_DEFAULT       = 1000
	DATE_START_DAY_KEY          = "day_start"
	DATE_DAYS_WINDOW_KEY        = "days_window"
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
	startingDay int
	howManyDays int
}

func githubCreationDay() time.Time {
	return time.Date(2007, time.November, 1, 12, 0, 0, 0, time.UTC)
}

func DefaultRepoCreationDateRange() RepoCreationDateRange {
	aboutToday := time.Now().AddDate(0, 0, 2)
	githubExistenceDays := int(aboutToday.Sub(githubCreationDay()).Hours() / 24.0)
	return RepoCreationDateRange{
		startingDay: 0,
		howManyDays: githubExistenceDays,
	}
}

func (r RepoCreationDateRange) CoversEverything() bool {
	return r.startingDay == DefaultRepoCreationDateRange().startingDay && r.CoversToday()
}

func GetRepoCreationDateRange(db *xorm.Engine) (r RepoCreationDateRange) {
	r.startingDay = getFromState[int](db, DATE_START_DAY_KEY, -1)
	if r.startingDay == -1 {
		return DefaultRepoCreationDateRange()
	}
	r.howManyDays = getFromState[int](db, DATE_DAYS_WINDOW_KEY, -1)
	if r.howManyDays == -1 {
		return DefaultRepoCreationDateRange()
	}
	return r
}

func (r RepoCreationDateRange) ToQueryString() string {
	start := githubCreationDay().AddDate(0, 0, r.startingDay)
	end := start.AddDate(0, 0, r.howManyDays)
	return fmt.Sprintf("%v..%v", start.Format("2006-01-02"), end.Format("2006-01-02"))
}

func (r RepoCreationDateRange) Save(db xorm.Interface) {
	// todo: transaction?
	setToState[int](db, DATE_START_DAY_KEY, r.startingDay)
	setToState[int](db, DATE_DAYS_WINDOW_KEY, r.howManyDays)
}

func (r RepoCreationDateRange) HalvedRange() (ret RepoCreationDateRange) {
	ret.startingDay = r.startingDay
	ret.howManyDays = int(smallerWindow(int64(r.howManyDays)))
	return ret
}

func (r RepoCreationDateRange) BiggerRange() (ret RepoCreationDateRange) {
	ret.startingDay = r.startingDay
	ret.howManyDays = int(biggerWindow(int64(r.howManyDays)))
	return ret
}

func (r RepoCreationDateRange) CoversToday() bool {
	end := githubCreationDay().
		AddDate(0, 0, r.startingDay).
		AddDate(0, 0, r.howManyDays)
	// add 24 hours just in case of either some one-of errors somewhere, or very new repos
	return end.After(time.Now().Add(24 * time.Hour))
}

func (r RepoCreationDateRange) NextRange() (ret RepoCreationDateRange) {
	ret.startingDay = r.startingDay + r.howManyDays + 1
	ret.howManyDays = r.howManyDays
	log.Printf("[creationDateRange] Going from %v to %v\n", r, ret)
	return ret
}
