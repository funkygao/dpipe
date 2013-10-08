package parser

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DbParser struct {
	DefaultParser
	db *sql.DB
}

// create table schema
func (this *DbParser) createDB(createTable string, dbFile string) {
	db, err := sql.Open(SQLITE3_DRIVER, dbFile)
	checkError(err)

	this.db = db

	stmt, err := this.db.Prepare(createTable)
	checkError(err)

	_, e := stmt.Exec()
	checkError(e)
}

func (this DbParser) insert(insertSql string, args ...interface {}) {
	stmt, err := this.db.Prepare(insertSql)
	checkError(err)

	_, e := stmt.Exec(args...)
	checkError(e)
}
