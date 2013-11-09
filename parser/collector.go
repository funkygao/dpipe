package parser

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/funkygao/alser/config"
	"github.com/funkygao/gotime"
	_ "github.com/mattn/go-sqlite3"
	"sync"
	"time"
)

// Child of AlsParser with db(sqlite3) features
type CollectorParser struct {
	AlsParser
	AlarmCollector

	*sync.Mutex

	db         *sql.DB
	insertStmt *sql.Stmt

	chWait  chan bool
	stopped bool
}

func (this *CollectorParser) init(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) {
	this.AlsParser.init(conf, chUpstream, chDownstream) // super

	this.Mutex = new(sync.Mutex) // embedding constructor
	this.chWait = make(chan bool)
	this.stopped = false

	this.createDB()
	this.prepareInsertStmt()
}

func (this *CollectorParser) Stop() {
	this.AlsParser.Stop() // super
	this.stopped = true

	if this.insertStmt != nil {
		this.insertStmt.Close()
	}
}

func (this *CollectorParser) Wait() {
	this.AlsParser.Wait()
	<-this.chWait

	if this.db != nil {
		this.db.Close()
	}
}

// TODO
// 各个字段显示顺心的问题，例如amount
// normalize
// payment的阶段汇总
// 有的字段需要运算，例如slowresp
// colorPrint的第一个字段必须是amount
func (this *CollectorParser) CollectAlarms() {
	if dryRun || !this.conf.Enabled {
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
		cols, _ := rows.Columns()
		count := len(cols)
		values := make([]interface{}, count)
		valuePtrs := make([]interface{}, count)
		mutex.Lock()
		this.echoCheckpoint(tsFrom, tsTo, this.conf.Title)
		var summary int64 = 0
		for rows.Next() {
			for i, _ := range cols {
				valuePtrs[i] = &values[i]
			}

			err := rows.Scan(valuePtrs...)
			checkError(this.id(), err)

			var amount = values[0].(int64)
			if amount == 0 {
				break
			}

			if this.conf.ShowSummary {
				summary += amount
			}

			if this.conf.BeepThreshold > 0 && int(amount) >= this.conf.BeepThreshold {
				this.beep()
				this.alarmf(this.conf.PrintFormat, values...)
			}

			this.colorPrintfLn(this.conf.PrintFormat, values...)
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
func (this *CollectorParser) createDB() {
	var err error
	this.db, err = sql.Open(SQLITE3_DRIVER, fmt.Sprintf("file:%s?cache=shared&mode=rwc",
		DATA_BASEDIR+this.conf.DbName+SQLITE3_DBFILE_SUFFIX))
	checkError(this.id(), err)

	_, err = this.db.Exec(fmt.Sprintf(this.conf.CreateTable, this.conf.DbName))
	checkError(this.id(), err)

	// performance tuning for sqlite3
	// http://www.sqlite.org/cvstrac/wiki?p=DatabaseIsLocked
	_, err = this.db.Exec("PRAGMA synchronous = OFF")
	checkError(this.id(), err)
	_, err = this.db.Exec("PRAGMA journal_mode = MEMORY")
	checkError(this.id(), err)
	_, err = this.db.Exec("PRAGMA read_uncommitted = true")
	checkError(this.id(), err)
}

func (this *CollectorParser) prepareInsertStmt() {
	if this.conf.InsertStmt == "" {
		panic("insert_stmt not configured")
	}

	var err error
	this.insertStmt, err = this.db.Prepare(fmt.Sprintf(this.conf.InsertStmt, this.conf.DbName))
	checkError(this.id(), err)
}

// auto lock/unlock
func (this *CollectorParser) insert(args ...interface{}) {
	this.Lock()
	_, err := this.insertStmt.Exec(args...)
	this.Unlock()
	checkError(this.id(), err)
}

// caller is responsible for locking
func (this *CollectorParser) execSql(sqlStmt string, args ...interface{}) (afftectedRows int64) {
	if debug {
		logger.Printf("%s %+v\n", sqlStmt, args)
	}

	res, err := this.db.Exec(sqlStmt, args...)
	checkError(this.id(), err)

	afftectedRows, err = res.RowsAffected()
	checkError(this.id(), err)

	return
}

func (this *CollectorParser) query(querySql string, args ...interface{}) *sql.Rows {
	if debug {
		logger.Printf("%s %+v\n", querySql, args)
	}

	rows, err := this.db.Query(querySql, args...)
	checkError(this.id(), err)

	return rows
}

// caller is responsible for locking
func (this *CollectorParser) delRecordsBefore(ts int) (affectedRows int64) {
	affectedRows = this.execSql("delete from "+this.conf.DbName+"  where ts<=?", ts)

	return
}

func (this *CollectorParser) getCheckpoint(wheres ...string) (tsFrom, tsTo int, err error) {
	query := fmt.Sprintf("SELECT min(ts), max(ts) FROM %s", this.conf.DbName)
	if len(wheres) > 0 {
		query += " WHERE 1=1"
		for _, w := range wheres {
			query += " AND " + w
		}
	}

	if debug {
		logger.Println(query)
	}

	row := this.db.QueryRow(query)
	err = row.Scan(&tsFrom, &tsTo)
	if err == nil && tsTo == 0 {
		err = errors.New("empty table")
	}

	if debug {
		logger.Println(tsFrom, tsTo, err)
	}

	return
}

func (this *CollectorParser) echoCheckpoint(tsFrom, tsTo int, title string) {
	fmt.Println() // seperator
	this.colorPrintfLn("(%s  ~  %s) %s", gotime.TsToString(tsFrom), gotime.TsToString(tsTo), title)
}
