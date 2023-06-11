package main

import (
	"net/url"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

func initialiseDb() *xorm.Engine {
	engine := lo.Must(xorm.NewEngine("sqlite3", (&url.URL{
		Path: *databasePath,
		RawQuery: url.Values{
			"mode":                {"rwc"},
			"_journal_mode":       {"WAL"},
			"_busy_timeout":       {"1000"},
			"_foreign_keys":       {"yes"},
			"_recursive_triggers": {"yes"},
			"_cache_size":         {"-32000"},
			"_synchronous":        {"NORMAL"},
			"_txlock":             {"exclusive"},
		}.Encode(),
	}).String()))

	engine.SetMapper(names.SameMapper{})

	if *enableSqlLog {
		sqlLog := lo.Must(os.OpenFile("logs/sql.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666))
		engine.SetLogger(log.NewSimpleLogger(sqlLog))
		engine.Logger().SetLevel(log.LOG_DEBUG)
		engine.ShowSQL(true)
	}

	// This creates a table if it doesn't exist, but doesn't update the schema it if differs from code
	lo.Must0(engine.Sync(
		new(Repo),
		new(State),
	))

	return engine
}
