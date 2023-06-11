package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"xorm.io/xorm"
)

func endWork(ctx context.Context, db *xorm.Engine) {
	SetMaxStars(db, MAX_STARS_DEFAULT)
	SetSearchWindow(db, SEARCH_WINDOW_DEFAULT)
	os.Exit(0)
}

func fetcherTask(ctx context.Context, db *xorm.Engine) {
	githubApiClient := newGithubApiClient(context.Background())
	for {
		maxStars := GetMaxStars(db)
		if maxStars < *minumumNumberOfStars {
			// We've downloaded everything there is
			log.Printf("MaxStars decreased to %v - ending all work.", maxStars)
			endWork(ctx, db)
			return
		}

		lo.TryCatchWithErrorValue(func() error {
			doFetcherTask(ctx, githubApiClient, db)
			return nil
		}, func(caught any) {
			type stackTracer interface{ StackTrace() errors.StackTrace }

			err, isAnActualError := caught.(error)
			stError, isStackTraceError := err.(stackTracer)

			if isAnActualError && isStackTraceError {
				log.Printf("Caught an error in fetcherTask: %+v\n%+v\n", caught, stError.StackTrace())
			} else {
				log.Printf("Caught an error in fetcherTask: %+v\n", caught)
			}
			log.Println("Will sleep for 30s and try again")
			time.Sleep(time.Second * 30)
		})
	}
}

func main() {
	if *enableRequestLog || *enableResponsesLog || *enableSqlLog {
		lo.Must0(os.MkdirAll("logs", os.ModePerm), "Couldn't mkdir -p ./logs/")
	}
	lo.Must0(os.MkdirAll("state", os.ModePerm), "Couldn't mkdir -p ./session_data/")

	initialiseGithubAppLogs()
	dbEngine := initialiseDb()
	defer dbEngine.Close()

	fetcherTask(context.Background(), dbEngine)

}
