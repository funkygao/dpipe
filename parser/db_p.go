package parser

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/funkygao/gotime"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

// Child of AlsParser with db(sqlite3) features
type DbParser struct {
	AlsParser
	AlarmCollector

	*sync.Mutex

	db         *sql.DB
	insertStmt *sql.Stmt

	dbName string // dbName IS table name for each db has only 1 table

	chWait chan bool
}

func (this *DbParser) init(name string, ch chan<- Alarm, dbFile, dbName, createTable, insertSql string) {
	this.AlsParser.init(name, ch) // super

	this.Mutex = new(sync.Mutex) // embedding constructor
	this.chWait = make(chan bool)
	this.dbName = dbName

	this.createDB(createTable, dbFile)
	this.prepareInsertStmt(insertSql)
}

func (this *DbParser) Stop() {
	this.AlsParser.Stop() // super
	if this.insertStmt != nil {
		this.insertStmt.Close()
	}
}

func (this *DbParser) Wait() {
	this.AlsParser.Wait()
	<-this.chWait

	if this.db != nil {
		this.db.Close()
	}
}

// create table schema
// for high TPS, each parser has a dedicated sqlite3 db file
func (this *DbParser) createDB(createTable string, dbFile string) {
	var err error
	this.db, err = sql.Open(SQLITE3_DRIVER, fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbFile))
	checkError(err)

	_, err = this.db.Exec(createTable)
	checkError(err)

	// performance tuning for sqlite3
	_, err = this.db.Exec("PRAGMA synchronous = OFF")
	checkError(err)
	_, err = this.db.Exec("PRAGMA journal_mode = MEMORY")
	checkError(err)
	_, err = this.db.Exec("PRAGMA read_uncommitted = true")
	checkError(err)
}

func (this *DbParser) prepareInsertStmt(insert string) {
	var err error
	this.insertStmt, err = this.db.Prepare(insert)
	checkError(err)
}

// auto lock/unlock
func (this *DbParser) insert(args ...interface{}) {
	this.Lock()
	_, err := this.insertStmt.Exec(args...)
	this.Unlock()
	checkError(err)
}

// caller is responsible for locking
func (this *DbParser) execSql(sqlStmt string, args ...interface{}) (afftectedRows int64) {
	res, err := this.db.Exec(sqlStmt, args...)
	checkError(err)

	afftectedRows, err = res.RowsAffected()
	checkError(err)

	return
}

func (this *DbParser) query(querySql string, args ...interface{}) *sql.Rows {
	rows, err := this.db.Query(querySql, args...)
	checkError(err)

	return rows
}

// caller is responsible for locking
func (this *DbParser) delRecordsBefore(ts int) (affectedRows int64) {
	affectedRows = this.execSql("delete from "+this.dbName+"  where ts<=?", ts)

	return
}

func (this *DbParser) checkpointSql(table string, wheres ...string) string {
	query := fmt.Sprintf("SELECT min(ts), max(ts) FROM %s", table)
	if len(wheres) > 0 {
		query += " WHERE 1=1"
		for _, w := range wheres {
			query += " AND " + w
		}
	}

	return query
}

func (this *DbParser) getCheckpoint(wheres ...string) (tsFrom, tsTo int, err error) {
	querySql := this.checkpointSql(this.dbName, wheres...)

	row := this.db.QueryRow(querySql)
	err = row.Scan(&tsFrom, &tsTo)
	if err == nil && tsTo == 0 {
		err = errors.New("empty table")
	}

	return
}

func (this *DbParser) logCheckpoint(color string, tsFrom, tsTo int, title string) {
	fmt.Println() // seperator
	this.colorPrintfLn(color, "(%s  ~  %s) %s", gotime.TsToString(tsFrom), gotime.TsToString(tsTo), title)
}
