package parser

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/funkygao/alser/config"
	"github.com/funkygao/gotime"
	_ "github.com/mattn/go-sqlite3"
	"strings"
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

	history map[string]int64 // TODO LRU incase of OOM

	chWait  chan bool
	stopped bool
}

func (this *CollectorParser) init(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) {
	this.AlsParser.init(conf, chUpstream, chDownstream) // super

	this.Mutex = new(sync.Mutex) // embedding constructor
	this.chWait = make(chan bool)
	this.stopped = false
	this.history = make(map[string]int64)

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

func (this *CollectorParser) isAbnormalChange(amount int64, key string) bool {
	defer func() {
		// will reset when history size is large enough
		if len(this.history) > 16384 { // (1<<20)/64
			this.history = make(map[string]int64) // clear
		}

		this.history[key] = amount // refresh
	}()

	if lastAmount, present := this.history[key]; present {
		delta := amount - lastAmount
		if delta < 0 {
			return false
		}

		if float64(delta)/float64(lastAmount) >= this.conf.AbormalPercent { // 20% by default
			return true
		}
	}

	return false
}

func (this *CollectorParser) historyKey(printf string, values []interface{}) string {
	parts := strings.SplitN(printf, "d", 2) // first column is always amount
	format := strings.TrimSpace(parts[1])
	key := fmt.Sprintf(format, values[1:]...) // raw key

	// use md5 to save memory
	h := md5.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

// TODO
// 有的字段需要运算，例如slowresp
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
			this.checkError(err)

			var amount = values[0].(int64)
			if amount == 0 {
				break
			}

			if this.conf.ShowSummary {
				summary += amount
			}

			abnormalChange := this.isAbnormalChange(amount, this.historyKey(this.conf.PrintFormat, values))
			shouldBeep := this.conf.BeepThreshold > 0 && int(amount) >= this.conf.BeepThreshold
			if shouldBeep || abnormalChange {
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
	this.checkError(err)

	_, err = this.db.Exec(fmt.Sprintf(this.conf.CreateTable, this.conf.DbName))
	this.checkError(err)

	// performance tuning for sqlite3
	// http://www.sqlite.org/cvstrac/wiki?p=DatabaseIsLocked
	_, err = this.db.Exec("PRAGMA synchronous = OFF")
	this.checkError(err)
	_, err = this.db.Exec("PRAGMA journal_mode = MEMORY")
	this.checkError(err)
	_, err = this.db.Exec("PRAGMA read_uncommitted = true")
	this.checkError(err)
}

func (this *CollectorParser) prepareInsertStmt() {
	if this.conf.InsertStmt == "" {
		panic("insert_stmt not configured")
	}

	var err error
	this.insertStmt, err = this.db.Prepare(fmt.Sprintf(this.conf.InsertStmt, this.conf.DbName))
	this.checkError(err)
}

// auto lock/unlock
func (this *CollectorParser) insert(args ...interface{}) {
	this.Lock()
	if debug {
		logger.Printf("%s %+v\n", this.id(), args)
	}
	_, err := this.insertStmt.Exec(args...)
	this.Unlock()
	this.checkError(err)
}

// caller is responsible for locking
func (this *CollectorParser) execSql(sqlStmt string, args ...interface{}) (afftectedRows int64) {
	if debug {
		logger.Printf("%s %+v\n", sqlStmt, args)
	}

	res, err := this.db.Exec(sqlStmt, args...)
	this.checkError(err)

	afftectedRows, err = res.RowsAffected()
	this.checkError(err)

	return
}

func (this *CollectorParser) query(querySql string, args ...interface{}) *sql.Rows {
	if debug {
		logger.Printf("%s %+v\n", querySql, args)
	}

	rows, err := this.db.Query(querySql, args...)
	this.checkError(err)

	return rows
}

// caller is responsible for locking
func (this *CollectorParser) delRecordsBefore(ts int) (affectedRows int64) {
	affectedRows = this.execSql("DELETE FROM "+this.conf.DbName+"  WHERE ts<=?", ts)

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
