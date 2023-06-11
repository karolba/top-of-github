package main

import "flag"

var enableRequestLog = flag.Bool("enable-request-log", false, "Log HTTP requests in ./logs/requests.log")
var enableResponsesLog = flag.Bool("enable-responses-log", false, "Log HTTP responses in ./logs/responses.log")
var enableSqlLog = flag.Bool("enable-sql-log", false, "Log SQL queries/statements in ./logs/sql.log")
var databasePath = flag.String("database", "state/repos.db", "Path to the sqlite database to use")
var minumumNumberOfStars = flag.Int64("minimum-stars", 10, "Metadata about repositories of this many stars and up will be downloaded")

func init() {
	flag.Parse()
}
