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
	this.mutexLock()

	res, err := this.db.Exec(sqlStmt, args...)
	checkError(err)

	afftectedRows, err = res.RowsAffected()
	checkError(err)

	this.mutexUnlock()
	return
}

func (this DbParser) query(querySql string, args ...interface{}) *sql.Rows {
	this.mutexLock()

	rows, err := this.db.Query(querySql, args...)
	checkError(err)

	this.mutexUnlock()

	return rows
}

func (this DbParser) getCheckpoint(querySql string, args ...interface{}) (ts int) {
	this.mutexLock()

	if err := this.db.QueryRow(querySql, args...).Scan(&ts); err != nil {
		ts = 0
	}

	this.mutexUnlock()

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
