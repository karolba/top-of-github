package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	ACCOUNT_ID        = haveToGetEnvironmentVariable("R2_ACCOUNT_ID")
	ACCESS_KEY_ID     = haveToGetEnvironmentVariable("R2_ACCESS_KEY_ID")
	ACCESS_KEY_SECRET = haveToGetEnvironmentVariable("R2_ACCESS_KEY_SECRET")
)

var (
	targetDirectory = flag.String("directory", ".", "A directory to upload")
	bucketName      = flag.String("bucket-name", "", "Target bucket name")
	contentType     = flag.String("content-type", "application/json", "Content-Type for uploaded files")
	contentEncoding = flag.String("content-encoding", "gzip", "Content-Type for uploaded files")
)

func init() {
	flag.Parse()

	if bucketName == nil || *bucketName == "" {
		fmt.Println("Error: The -bucket-name parameter is mandatory")
		os.Exit(1)
	}
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
