package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"time"
)

// Errlog's MongoException parser
type MongodbLogParser struct {
	DbParser
}

// Constructor
func newMongodbLogParser(name, color string, chAlarm chan<- Alarm, dbFile, dbName, createTable, insertSql string) (parser *MongodbLogParser) {
	parser = new(MongodbLogParser)
	parser.init(name, color, chAlarm, dbFile, dbName, createTable, insertSql)

	go parser.CollectAlarms()

	return
}

func (this *MongodbLogParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DbParser.ParseLine(line)
	if dryRun {
		return
	}

	cls, err := data.Get("class").String()
	if err != nil || cls != "MongoException" {
		// not a mongodb log
		return
	}

	level, err := data.Get("level").String()
	checkError(err)
	msg, err := data.Get("message").String()
	checkError(err)
	msg = this.normalizeMsg(msg)
	flash, err := data.Get("flash_version_client").String()

	logInfo := extractLogInfo(data)
	this.insert(area, ts, level, msg, flash, logInfo.host)

	return
}

func (this *MongodbLogParser) normalizeMsg(msg string) string {
	r := digitsRegexp.ReplaceAll([]byte(msg), []byte("?"))
	return string(r)
}

func (this *MongodbLogParser) CollectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	sleepInterval := time.Duration(this.conf.Int("sleep", 15))
	beepThreshold := this.conf.Int("beep_threshold", 1)

	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select count(*) as am, msg from mongo where ts<=? group by msg order by am desc", tsTo)
		parsersLock.Lock()
		this.echoCheckpoint(tsFrom, tsTo, "MongoException")
		for rows.Next() {
			var msg string
			var amount int64
			err := rows.Scan(&amount, &msg)
			checkError(err)

			if amount >= int64(beepThreshold) {
				this.beep()
				this.alarmParserPrintf("%5s %s", gofmt.Comma(amount), msg)
			}

			this.colorPrintfLn("%5s %s", gofmt.Comma(amount), msg)
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
