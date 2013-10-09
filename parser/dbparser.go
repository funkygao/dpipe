package parser

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"sync"
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
	db, err := sql.Open(SQLITE3_DRIVER, dbFile)
	checkError(err)

	this.db = db
	this.lock = new(sync.Mutex)

	stmt, err := this.db.Prepare(createTable)
	checkError(err)

	_, e := stmt.Exec()
	checkError(e)
}

func (this DbParser) execSql(sqlStmt string, args ...interface{}) (afftectedRows int64) {
	this.mutexLock()
	defer this.mutexUnlock()

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
	this.mutexLock()
	stmt, err := this.db.Prepare(querySql)
	checkError(err)

	if err := stmt.QueryRow(args...).Scan(&ts); err != nil {
		ts = 0
	}
	this.mutexUnlock()

	return
}

func (this DbParser) logCheckpoint(ts int) {
	jst, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	bjtime := t.In(jst)
	logger.Println(bjtime)
}
