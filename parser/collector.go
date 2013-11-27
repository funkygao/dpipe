package parser

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/funkygao/alser/config"
	sqldb "github.com/funkygao/alser/db"
	"github.com/funkygao/gotime"
	"os"
	"strings"
	"sync"
	"time"
)

// Child of AlsParser with db(sqlite3) features
type CollectorParser struct {
	AlsParser
	AlarmCollector

	*sync.Mutex

	db          *sqldb.SqlDb
	pdb         *sqldb.SqlDb // persist db
	insertStmt  *sql.Stmt
	pinsertStmt *sql.Stmt // persist
	statsStmt   *sql.Stmt

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
	this.prepareStatsStmt()
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
		if len(this.history) > (5<<20)/64 {
			// each parser consumes 5M history data
			// each history entry consumes 64bytes
			this.history = make(map[string]int64)
			logger.Printf("[%s] history data cleared\n", this.id())
		}

		this.history[key] = amount // refresh
	}()

	if lastAmount, present := this.history[key]; present {
		delta := amount - lastAmount
		if delta < 0 {
			return false
		}

		if float64(delta)/float64(lastAmount) >= this.conf.AbnormalPercent {
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

func (this *CollectorParser) CollectAlarms() {
	if dryRun || !this.conf.Enabled {
		this.chWait <- true
		return
	}

	for {
		time.Sleep(time.Second * time.Duration(this.conf.Sleep))

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows, _ := this.statsStmt.Query(tsTo)
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

			if this.conf.BeepThreshold > 0 && int(amount) >= this.conf.BeepThreshold {
				this.beep()
				this.alarmf(this.conf.PrintFormat, values...)
			}

			if amount >= int64(this.conf.AbnormalBase) &&
				this.isAbnormalChange(amount, this.historyKey(this.conf.PrintFormat, values)) {
				this.beep()
				this.blinkColorPrintfLn(this.conf.PrintFormat, values...)
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
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc",
		fmt.Sprintf("%s/%s-%d.%s", DATA_BASEDIR, this.conf.DbName, os.Getpid(),
			SQLITE3_DBFILE_SUFFIX))
	this.db = sqldb.NewSqlDb(sqldb.DRIVER_SQLITE3, dsn, logger)
	this.db.SetDebug(debug)
	this.db.CreateDb(fmt.Sprintf(this.conf.CreateTable, this.conf.DbName))

	if this.conf.PersistDb != "" {
		dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc",
			fmt.Sprintf("%s/%s.%s", DATA_BASEDIR, this.conf.PersistDb,
				PERSIST_DBFILE_SUFFIX))
		this.pdb = sqldb.NewSqlDb(sqldb.DRIVER_SQLITE3, dsn, logger)
		this.pdb.SetDebug(debug)
		this.pdb.CreateDb(fmt.Sprintf(this.conf.CreateTable, this.conf.PersistDb))
	}
}

func (this *CollectorParser) prepareInsertStmt() {
	if this.conf.InsertStmt == "" {
		panic("insert_stmt not configured")
	}

	this.insertStmt = this.db.Prepare(fmt.Sprintf(this.conf.InsertStmt, this.conf.DbName))
	if this.pdb != nil {
		this.pinsertStmt = this.pdb.Prepare(fmt.Sprintf(this.conf.InsertStmt, this.conf.PersistDb))
	}
}

func (this *CollectorParser) prepareStatsStmt() {
	statsSql := this.conf.StatsSql()
	if statsSql == "" {
		panic("stats_stmt not configured")
	}

	this.statsStmt = this.db.Prepare(statsSql)
}

// auto lock/unlock
func (this *CollectorParser) insert(args ...interface{}) {
	if debug {
		logger.Printf("%s %+v\n", this.id(), args)
	}

	this.Lock()
	_, err := this.insertStmt.Exec(args...)
	this.Unlock()
	this.checkError(err)

	if this.pinsertStmt != nil {
		this.pinsertStmt.Exec(args...)
	}
}

// caller is responsible for locking
func (this *CollectorParser) delRecordsBefore(ts int) (affectedRows int64) {
	affectedRows = this.db.ExecSql("DELETE FROM "+this.conf.DbName+"  WHERE ts<=?", ts)

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
