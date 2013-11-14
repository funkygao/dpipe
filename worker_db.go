package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
	sqldb "github.com/funkygao/alser/db"
	"github.com/funkygao/alser/parser"
	"strings"
	"sync"
	"time"
)

/*
+-------------+---------------------+------+-----+---------+----------------+
| Field       | Type                | Null | Key | Default | Extra          |
+-------------+---------------------+------+-----+---------+----------------+
| id          | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
| uid         | bigint(20) unsigned | NO   | MUL | NULL    |                |
| type        | int(10) unsigned    | NO   | MUL | NULL    |                |
| data        | blob                | NO   |     | NULL    |                |
| ip          | bigint(20)          | NO   | MUL | NULL    |                |
| ua          | int(10) unsigned    | NO   | MUL | NULL    |                |
| date_create | int(10) unsigned    | NO   | MUL | NULL    |                |
+-------------+---------------------+------+-----+---------+----------------+
*/
type DbWorker struct {
	Worker
	Lines chan string
	db    *sqldb.SqlDb
}

func newDbWorker(id int,
	dataSource string, conf config.ConfGuard, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int, chAlarm chan<- parser.Alarm) Runnable {
	this := new(DbWorker)
	this.Worker = Worker{id: id,
		dataSource: dataSource, conf: conf, tailMode: tailMode,
		wg: wg, Mutex: mutex,
		chLines: chLines, chAlarm: chAlarm}
	this.Lines = make(chan string)
	this.db = sqldb.NewSqlDb(sqldb.DRIVER_MYSQL, FLASHLOG_DSN, logger)
	this.db.Debug(options.debug)
	return this
}

func (this *DbWorker) Run() {
	defer this.Done()

	go this.feedLines()

	for line := range this.Lines {
		// a valid line scanned
		this.chLines <- 1

		// feed the parsers one by one
		for _, parserId := range this.conf.Parsers {
			parser.Dispatch(parserId, line)
		}
	}

	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}

}

func (this *DbWorker) feedLines() {
	var lastId int64 = this.getLastId()
	if lastId < 0 {
		logger.Printf("table[%s] skipped\n", this.dataSource)

		close(this.Lines)
		return
	}

	for {
		time.Sleep(20)

		var (
			id   int64
			typ  int
			data string
		)

		rows := this.db.Query(fmt.Sprintf("SELECT id,type,data FROM %s WHERE id>=%d ORDER BY id", this.dataSource, lastId))
		for rows.Next() {
			if err := rows.Scan(&id, &typ, &data); err != nil {
				panic(err)
			}

			if line := this.genLine(typ, data); line != "" {
				this.Lines <- line
			}
		}

		lastId = id
	}
}

func (this *DbWorker) getLastId() (lastId int64) {
	row := this.db.QueryRow(fmt.Sprintf("SELECT max(id) from %s", this.dataSource))
	if err := row.Scan(&lastId); err != nil {
		if options.verbose || options.debug {
			logger.Printf("%s %s\n", this.dataSource, err.Error())
		}

		lastId = -1
	}

	return
}

func (this *DbWorker) area() string {
	p := strings.SplitN(this.dataSource, "_", 2)
	return p[1]
}

func (this *DbWorker) genLine(typ int, data string) string {
	// gzuncompress data
	r, err := zlib.NewReader(bytes.NewBufferString(data))
	if err != nil {
		panic(err)
	}
	defer r.Close()

	var d []byte
	if _, err := r.Read(d); err != nil {
		panic(err)
	}

	var jsonData *json.Json
	var e error
	if jsonData, e = json.NewJson(d); e != nil {
		logger.Printf("unkown feed: %s\n", data)
		return ""
	}

	logger.Printf("%s\n%v", string(d), *jsonData)

	return ""
}
