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

func rowAsRecord(singleRow *sql.Rows, columns []string) Record {
	record := make(Record)
	values := make([]any, len(columns))

	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	if err := singleRow.Scan(valuePtrs...); err != nil {
		log.Fatalln(err)
	}

	for i, col := range columns {
		record[col] = values[i]
	}

	return record
}

func rowsAsRecords(rows *sql.Rows, columns []string) []Record {
	records := []Record{}
	for rows.Next() {
		records = append(records, rowAsRecord(rows, columns))
	}
	return records
}
