package parser

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/funkygao/alser/config"
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

	chWait  chan bool
	stopped bool
}

func (this *DbParser) init(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) {
	this.AlsParser.init(name, color, ch) // super

	this.Mutex = new(sync.Mutex) // embedding constructor
	this.chWait = make(chan bool)
	this.stopped = false

	this.createDB()
	this.prepareInsertStmt()
}

func (this *DbParser) Stop() {
	this.AlsParser.Stop() // super
	this.stopped = true

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

func (this *DbParser) CollectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	statsSql := this.conf.StatsSql()

	for {
		time.Sleep(time.Second * time.Duration(this.conf.Sleep))

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query(this.conf.StatsSql(), tsTo)
		parsersLock.Lock()
		this.echoCheckpoint(tsFrom, tsTo, this.conf.Title)
		for rows.Next() {
			cols, _ = rows.Columns()
			pointers := make([]interface{}, len(cols)-1)
			container := make([]sql.NullString, len(cols)-1)
			for i, _ := range cols[1:] {
				pointers[i] = &container[i]
			}
			var amount int64
			err := rows.Scan(&amount, pointers...)
			checkError(err)

			if int(amount) >= this.conf.BeepThreshold {
				this.beep()
				this.alarmf(this.conf.PrintFormat, pointers...)
			}

			this.colorPrintfLn(this.conf.PrintFormat, pointers)
		}
		parsersLock.Unlock()
		rows.Close()

		this.delRecordsBefore(tsTo)
		this.Unlock()

		if this.stopped {
			this.chWait <- true
			break
		}
	}
}

// create table schema
// for high TPS, each parser has a dedicated sqlite3 db file
func (this *DbParser) createDB() {
	var err error
	this.db, err = sql.Open(SQLITE3_DRIVER, fmt.Sprintf("file:%s?cache=shared&mode=rwc", this.conf.DbName))
	checkError(err)

	_, err = this.db.Exec(this.conf.CreateTable)
	checkError(err)

	// performance tuning for sqlite3
	// http://www.sqlite.org/cvstrac/wiki?p=DatabaseIsLocked
	_, err = this.db.Exec("PRAGMA synchronous = OFF")
	checkError(err)
	_, err = this.db.Exec("PRAGMA journal_mode = MEMORY")
	checkError(err)
	_, err = this.db.Exec("PRAGMA read_uncommitted = true")
	checkError(err)
}

func (this *DbParser) prepareInsertStmt() {
	if this.conf.InsertStmt == "" {
		panic("insert_stmt not configured")
	}

	var err error
	this.insertStmt, err = this.db.Prepare(this.conf.InsertStmt)
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
	affectedRows = this.execSql("delete from "+this.conf.DbName+"  where ts<=?", ts)

	return
}

func (this *DbParser) getCheckpoint(wheres ...string) (tsFrom, tsTo int, err error) {
	query := fmt.Sprintf("SELECT min(ts), max(ts) FROM %s", this.conf.DbName)
	if len(wheres) > 0 {
		query += " WHERE 1=1"
		for _, w := range wheres {
			query += " AND " + w
		}
	}

	row := this.db.QueryRow(query)
	err = row.Scan(&tsFrom, &tsTo)
	if err == nil && tsTo == 0 {
		err = errors.New("empty table")
	}

	return
}

func (this *DbParser) echoCheckpoint(tsFrom, tsTo int, title string) {
	fmt.Println() // seperator
	this.colorPrintfLn("(%s  ~  %s) %s", gotime.TsToString(tsFrom), gotime.TsToString(tsTo), title)
}
