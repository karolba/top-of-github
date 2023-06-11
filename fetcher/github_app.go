package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/beatlabs/github-auth/app"
	"github.com/beatlabs/github-auth/key"
	"github.com/samber/lo"
)

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

var GITHUB_APP_APP_ID = haveToGetEnvironmentVariable("GITHUB_APP_APP_ID")
var GITHUB_APP_INSTALLATION_ID = haveToGetEnvironmentVariable("GITHUB_APP_INSTALLATION_ID")
var GITHUB_APP_PRIVATE_KEY_PEM_FILE_PATH = haveToGetEnvironmentVariable("GITHUB_APP_PRIVATE_KEY_PEM_FILE_PATH")
var githubApiPrivateKeyPem []byte

func init() {
	githubApiPrivateKeyPem = lo.Must(os.ReadFile(GITHUB_APP_PRIVATE_KEY_PEM_FILE_PATH))
}

const MAX_RESULTS_PER_PAGE = 100
const MAX_PAGES = 10

var reqLogger *log.Logger
var resLogger *log.Logger

func initialiseGithubAppLogs() {
	if *enableRequestLog {
		reqLogFile := lo.Must(os.OpenFile("logs/requests.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666))
		reqLogger = log.New(reqLogFile, "\n[http-request] ", log.Flags())
	}

	if *enableResponsesLog {
		resLogFile := lo.Must(os.OpenFile("logs/responses.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666))
		resLogger = log.New(resLogFile, "\n[http-response] ", log.Flags())
	}
}

func newGithubApiClient(ctx context.Context) *http.Client {
	ghApiPrivateKey := lo.Must(key.Parse(githubApiPrivateKeyPem))
	appConfig := lo.Must(app.NewConfig(GITHUB_APP_APP_ID, ghApiPrivateKey))
	installationConfig := lo.Must(appConfig.InstallationConfig(GITHUB_APP_INSTALLATION_ID))
	return installationConfig.Client(ctx)
}
