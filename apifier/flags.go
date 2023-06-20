package main

import "flag"

var (
	outputDir    = flag.String("output-dir", ".", "Where to save generated json files")
	databasePath = flag.String("database", "state/repos.db", "Path to the sqlite database to use")
)

func init() {
	flag.Parse()
}
