package main

import (
	"database/sql"
	"fmt"
	"log"
)

// Retrieve column names from the SQLite table
func getColumnNames(db *sql.DB, tableName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer closeOrPanic(rows)

	columns := []string{}
	for rows.Next() {
		var cid int
		var name string
		var _type string
		var notnull int
		var dfltValue any
		var pk int
		err := rows.Scan(&cid, &name, &_type, &notnull, &dfltValue, &pk)
		if err != nil {
			return nil, err
		}
		columns = append(columns, name)

		if err := rows.Err(); err != nil {
			return nil, err
		}

	}
	return columns, nil
}

func rowsAsRecords(rows *sql.Rows, columns []string) []Record {
	// Create a slice to hold the records
	records := []Record{}

	// Iterate over the rows and store the data
	for rows.Next() {
		record := make(Record)
		// Create a slice to hold the values of each column
		values := make([]any, len(columns))
		// Create a slice to hold the pointers to the values
		valuePtrs := make([]any, len(columns))

		// Populate the value pointers with the addresses of the corresponding column values
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Fatalln(err)
		}

		// Map column names to their values in the record
		for i, col := range columns {
			record[col] = values[i]
		}

		records = append(records, record)
	}
	return records
}
