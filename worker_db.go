package main

import (
	"database/sql"
	"fmt"
	"github.com/funkygao/alser/config"
	sqldb "github.com/funkygao/alser/db"
	"github.com/funkygao/alser/parser"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type DbWorker struct {
	Worker
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
	return this
}

func (this *DbWorker) Run() {
	defer this.Done()

}

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
func flashlogDataSource() {
	db, err := sql.Open(sqldb.DRIVER_MYSQL, FLASHLOG_DSN)
	if err != nil {
		panic(err)
		return
	}

	rows, err := db.Query("select * from log_us WHERE type=299 ORDER BY ID limit 10")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		fmt.Println(rows)
	}
}
