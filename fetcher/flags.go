package main

import "flag"

var (
	enableRequestLog     = flag.Bool("enable-request-log", false, "Log HTTP requests in ./logs/requests.log")
	enableResponsesLog   = flag.Bool("enable-responses-log", false, "Log HTTP responses in ./logs/responses.log")
	enableSqlLog         = flag.Bool("enable-sql-log", false, "Log SQL queries/statements in ./logs/sql.log")
	databasePath         = flag.String("database", "state/repos.db", "Path to the sqlite database to use")
	minumumNumberOfStars = flag.Int64("minimum-stars", 5, "Metadata about repositories of this many stars and up will be downloaded")
)

func init() {
	flag.Parse()
}
