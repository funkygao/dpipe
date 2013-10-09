package parser

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
)

type DbParser struct {
	DefaultParser
	db   *sql.DB
	lock *sync.Mutex
}

func (this *DbParser) mutexLock() {
	this.lock.Lock()
}

func (this *DbParser) mutexUnlock() {
	this.lock.Unlock()
}

// create table schema
func (this *DbParser) createDB(createTable string, dbFile string) {
	db, err := sql.Open(SQLITE3_DRIVER, fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbFile))
	checkError(err)

	this.db = db
	this.lock = new(sync.Mutex)

	_, err = this.db.Exec(createTable)
	checkError(err)
}

func (this DbParser) execSql(sqlStmt string, args ...interface{}) (afftectedRows int64) {
	stmt, err := this.db.Prepare(sqlStmt)
	checkError(err)

	res, err := stmt.Exec(args...)
	checkError(err)

	afftectedRows, err = res.RowsAffected()
	checkError(err)

	return
}

func (this DbParser) query(querySql string, args ...interface{}) *sql.Rows {
	rows, err := this.db.Query(querySql, args...)
	checkError(err)

	return rows
}

func (this DbParser) getCheckpoint(querySql string, args ...interface{}) (ts int) {
	stmt, err := this.db.Prepare(querySql)
	checkError(err)

	if err := stmt.QueryRow(args...).Scan(&ts); err != nil {
		ts = 0
	}

	return
}

func (this DbParser) logCheckpoint(ts int) {
	t := time.Unix(int64(ts), 0)
	jst, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	bjtime := t.In(jst)
	logger.Println(bjtime)
}
