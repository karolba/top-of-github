package main

import (
	"flag"
	"log"
	"os"
)

var (
	ACCOUNT_ID        = haveToGetEnvironmentVariable("R2_ACCOUNT_ID")
	ACCESS_KEY_ID     = haveToGetEnvironmentVariable("R2_ACCESS_KEY_ID")
	ACCESS_KEY_SECRET = haveToGetEnvironmentVariable("R2_ACCESS_KEY_SECRET")
	BUCKET_NAME       = haveToGetEnvironmentVariable("R2_BUCKET_NAME")
)

var targetDirectory = flag.String("directory", ".", "A directory to upload")

func init() {
	flag.Parse()
}

func haveToGetEnvironmentVariable(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Panicf("Missing required environment variable %s\n", name)
	}
	if val == "" {
		log.Panicf("Required environment variable %s is empty\n", name)
	}
	return val
}
