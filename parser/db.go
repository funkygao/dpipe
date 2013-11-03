package parser

import (
	"database/sql"
	"errors"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
	"github.com/funkygao/gotime"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
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

func newDbParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *DbParser) {
	this = new(DbParser)
	this.init(conf, chUpstream, chDownstream)

	go this.CollectAlarms()

	return
}

func (this *DbParser) init(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) {
	this.AlsParser.init(conf, chUpstream, chDownstream) // super

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

func (this *DbParser) ParseLine(line string) (area string, ts uint64, msg string) {
	var data *json.Json
	area, ts, data = this.AlsParser.parseJsonLine(line)
	if dryRun {
		return
	}

	args, err := this.extractKeyValues(data)
	if err != nil {
		return
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	//this.insert(area, ts, args...)

	return
}

// TODO
// 各个字段显示顺心的问题，例如amount
// normalize
// payment的阶段汇总
// 有的字段需要运算，例如slowresp
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

		rows := this.query(statsSql, tsTo)
		mutex.Lock()
		this.echoCheckpoint(tsFrom, tsTo, this.conf.Title)
		var summary int = 0
		for rows.Next() {
			cols, _ := rows.Columns()
			pointers := make([]interface{}, len(cols))
			container := make([]sql.NullString, len(cols))
			for i, _ := range cols {
				pointers[i] = &container[i]
			}

			err := rows.Scan(pointers...)
			checkError(err)

			var amount = pointers[0].(int)
			if amount == 0 {
				break
			}

			if this.conf.ShowSummary {
				summary += amount
			}

			if this.conf.BeepThreshold > 0 && amount >= this.conf.BeepThreshold {
				this.beep()
				this.alarmf(this.conf.PrintFormat, pointers...)
			}

			this.colorPrintfLn(this.conf.PrintFormat, pointers)
		}

		if this.conf.ShowSummary && summary > 0 {
			this.colorPrintfLn("Total: %d", summary)
		}
		mutex.Unlock()
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

	_, err = this.db.Exec(fmt.Sprintf(this.conf.CreateTable, this.conf.DbName))
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
	this.insertStmt, err = this.db.Prepare(fmt.Sprintf(this.conf.InsertStmt, this.conf.DbName))
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
