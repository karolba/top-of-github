package main

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"code.gitea.io/gitea/modules/emoji"
	"github.com/leporo/sqlf"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var fileSaveWaitGroup sync.WaitGroup

const JSON_PAGINATION_PAGE_SIZE = 500

const MINIMUM_REPOSITORY_STARGAZERS = 5

type Record map[string]any

type closable interface {
	Close() error
}

func closeOrPanic(toClose closable) {
	err := toClose.Close()
	if err != nil {
		log.Fatalln("Could not call .Close(): ", err)
	}
}

func programmingLanguages() []string {
	// Query to retrieve unique languages
	// This query is done on "Repo" instead of "ActiveRepo" to be sure to override the result json file for every
	// programming language we've ever came across.
	query := "SELECT DISTINCT Language FROM Repo"

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalln("Error executing query:", err)
	}
	defer closeOrPanic(rows)

	var languages []string
	for rows.Next() {
		var language string
		if err := rows.Scan(&language); err != nil {
			log.Fatalln("Error scanning row:", err)
		}
		languages = append(languages, language)
	}
	return languages
}

// createView creates a view of repos with fields actually used on the frontend.
// additionally, the repos have to have been found at least 15 search cycles back
// - this prevents displaying deleted repositories.
func createActiveRepoView() {
	// Have to use sprintf here, because as sqlite points out: "parameters are not allowed in views"
	_, err := db.Exec(fmt.Sprintf(`
		begin transaction;
		drop view if exists ActiveRepo;
		create view ActiveRepo as
		    select
		        Id,
		        Archived,
		        CreatedAt,
		        Description,
		        GithubLink,
		        Homepage,
		        Language,
		        LicenseSpdxId,
		        LicenseName,
		        Name,
		        OwnerAvatarUrl,
		        OwnerLogin,
		        RepoPushedAt,
		        RepoUpdatedAt,
		        Stargazers
		    from Repo
		    where Repo.Stargazers >= %d;
		end;
	`, MINIMUM_REPOSITORY_STARGAZERS))
	if err != nil {
		log.Fatalln("Could not create the ActiveRepo view:", err)
	}
}

func createIndices() {
	log.Print("Creating index on Repo(Language, Stargazers, Id, NotSeenSinceCounter)... ")
	_, err := db.Exec(`
		create index if not exists LanguageStargazersId on Repo(Language, Stargazers DESC, Id, NotSeenSinceCounter);
	`)
	if err != nil {
		log.Fatalln("\nCould not create index LanguageStargazers:", err)
	}
	log.Println("done")

	log.Print("Creating index on Repo(Stargazers, Id, NotSeenSinceCounter)... ")
	_, err = db.Exec(`
		create index if not exists StargazersId on Repo(Stargazers DESC, Id, NotSeenSinceCounter);
	`)
	if err != nil {
		log.Fatalln("\nCould not create index Stargazers:", err)
	}
	log.Println("done")
}

func dropIndices() {
	log.Print("Dropping index on Repo(Language, Stargazers, Id)... ")
	_, err := db.Exec(`drop index LanguageStargazersId;`)
	if err != nil {
		log.Fatalln("\nCould not drop index LanguageStargazersId", err)
	}
	log.Println("done")

	log.Print("Dropping index on Repo(Stargazers, Id)... ")
	_, err = db.Exec(`drop index StargazersId;`)
	if err != nil {
		log.Fatalln("\nCould not drop index StargazersId", err)
	}
	log.Println("done")
}

func escapeLanguageName(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "&", "-")
	name = strings.ReplaceAll(name, "?", "-")
	name = strings.ReplaceAll(name, "#", "-sharp-")
	if name == "" {
		return "-empty-"
	}
	return name
}

func main() {
	// Open a connection to the SQLite database
	var err error
	db, err = sql.Open("sqlite3", fmt.Sprintf("%s?mode=rw&_busy_timeout=-5000&_journal_mode=WAL", *databasePath))
	if err != nil {
		log.Fatalln(err)
	}
	defer closeOrPanic(db)

	createActiveRepoView()
	createIndices()
	defer dropIndices()

	// Retrieve column names from the table
	columnNames, err := getColumnNames(db, "ActiveRepo")
	if err != nil {
		log.Fatalln(err)
	}

	saveMetadata()

	// Retrieve all possible languages from the Repo table
	languages := programmingLanguages()

	// A hack: GitHub tags vimscript as three separate language names: "Vim Script", "Vim script", and "VimL"
	// - pretend they're the same thing
	githubLanguageNamesForVimScript := []string{"Vim Script", "Vim script", "VimL"}
	exportForLanguage("Vim Script / VimL", githubLanguageNamesForVimScript, columnNames)

	for _, language := range languages {
		if slices.Contains(githubLanguageNamesForVimScript, language) {
			break
		}
		exportForLanguage(language, []string{language}, columnNames)
	}

	exportForAll(columnNames)

	fileSaveWaitGroup.Wait()
}

func exportForAll(columnNames []string) {
	// Set the page size and initialize the offset
	pageSize := JSON_PAGINATION_PAGE_SIZE
	offset := 0
	page := 1

	for retrieveAndSaveAll(columnNames, pageSize, offset, page) {
		// Update offset and page number
		offset += pageSize
		page++
	}
}

func exportForLanguage(language string, githubNamesForTheLanguage []string, columnNames []string) {
	// Set the page size and initialize the offset
	pageSize := JSON_PAGINATION_PAGE_SIZE
	offset := 0
	page := 1

	for retrieveAndSaveByLanguage(columnNames, pageSize, offset, page, language, githubNamesForTheLanguage) {
		// Update offset and page number
		offset += pageSize
		page++
	}
}

// emojify renders all repository description emojis into unicode emojis
// For example: turns Description=":rocket: LGTM" into Description="🚀 LGTM"
func emojify(records []Record) []Record {
	for i, record := range records {
		description, ok := record["Description"]
		if !ok {
			continue
		}

		desc, ok := description.(string)
		if !ok {
			continue
		}

		records[i]["Description"] = emoji.ReplaceAliases(desc)
	}
	return records
}

func retrieveAndSaveAll(columnNames []string, pageSize int, offset int, page int) (shouldContinue bool) {
	// Retrieve data from the database with pagination
	rows, err := db.Query(`
		SELECT * FROM ActiveRepo
		ORDER BY Stargazers DESC, Id
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		log.Fatalln(err)
	}
	defer closeOrPanic(rows)

	fileName := fmt.Sprintf("%s/all/%d", *outputDir, page)

	records := rowsAsRecords(rows, columnNames)
	records = emojify(records)

	fileSaveWaitGroup.Add(1)
	go saveToFile(fileName, records)

	// Break the loop if there are no more records
	shouldContinue = len(records) >= pageSize

	return shouldContinue
}

func stringSliceToAnySlice(s []string) []any {
	ret := make([]interface{}, len(s))
	for i := range s {
		ret[i] = s[i]
	}
	return ret
}

func retrieveAndSaveByLanguage(columnNames []string, pageSize int, offset int, page int, language string, githubNamesForTheLanguage []string) (shouldContinue bool) {
	fileName := fmt.Sprintf("%s/language/%s/%d", *outputDir, escapeLanguageName(language), page)

	records := make([]Record, 0, pageSize)

	err := sqlf.From("ActiveRepo").
		Select("*").
		Where("Language").
		In(stringSliceToAnySlice(githubNamesForTheLanguage)...).
		OrderBy("Stargazers DESC, Id").
		Limit(pageSize).
		Offset(offset).
		QueryAndClose(context.Background(), db, func(row *sql.Rows) {
			records = append(records, rowAsRecord(row, columnNames))
		})
	if err != nil {
		log.Fatalln(err)
	}

	records = emojify(records)

	fileSaveWaitGroup.Add(1)
	go saveToFile(fileName, records)

	// Break the loop if there are no more records
	shouldContinue = len(records) >= pageSize

	return shouldContinue
}

func saveToFile(fileName string, records []Record) {
	// Convert records to JSON
	jsonData, err := json.Marshal(records)
	if err != nil {
		log.Fatalln(err)
	}

	// Write JSON data to a file
	saveDataToGzipFile(fileName, jsonData)
}

func saveDataToGzipFile(fileName string, data []byte) {
	defer fileSaveWaitGroup.Done()

	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create directory %v: %v\n", filepath.Dir(fileName), err)
	}

	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer closeOrPanic(file)

	gzipWriter := gzip.NewWriter(file)
	defer closeOrPanic(gzipWriter)

	// Write data to the gzip file
	_, err = gzipWriter.Write(data)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Created file '%s'\n", fileName)
}
