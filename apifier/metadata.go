package main

import (
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Language struct {
	Name         string
	EscapedName  string
	CountOfRepos int64
	CountOfStars int64
}

type Metadata struct {
	CountOfAllRepos int64
	CountOfAllStars int64
	Languages       []Language
}

func saveMetadata() {
	// Query for count of all repos
	var countOfAllRepos int64
	err := db.QueryRow("SELECT COUNT(*) FROM Repo").Scan(&countOfAllRepos)
	if err != nil {
		log.Fatal(err)
	}

	// Query for count of all stars
	var countOfAllStars int64
	err = db.QueryRow("SELECT SUM(Stargazers) FROM Repo").Scan(&countOfAllStars)
	if err != nil {
		log.Fatal(err)
	}

	// Query for count of repos and stars per language
	rows, err := db.Query(`
		SELECT Language, SUM(Stargazers), COUNT(*)
		FROM Repo
		GROUP BY Language
		ORDER BY COUNT(*) DESC, Name
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer closeOrPanic(rows)

	// Store the languages and their counts in a slice
	var languages []Language
	for rows.Next() {
		var languageName string
		var countOfStars int64
		var countOfRepos int64
		err := rows.Scan(&languageName, &countOfStars, &countOfRepos)
		if err != nil {
			log.Fatal(err)
		}
		languages = append(languages, Language{
			Name:         languageName,
			EscapedName:  escapeLanguageName(languageName),
			CountOfRepos: countOfRepos,
			CountOfStars: countOfStars,
		})
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Create the Metadata struct and populate it with the extracted data
	data := Metadata{
		CountOfAllRepos: countOfAllRepos,
		CountOfAllStars: countOfAllStars,
		Languages:       languages,
	}

	// Marshal the data to JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	go saveDataToGzipFile(fmt.Sprintf("%s/metadata", *outputDir), jsonData)
}
