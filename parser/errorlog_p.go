package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"strings"
	"time"
)

// Errlog parser
type ErrorLogParser struct {
	DbParser
	skippedErrors []string
}

// Constructor
func newErrorLogParser(name string, chAlarm chan<- Alarm, dbFile, dbName, createTable, insertSql string) (parser *ErrorLogParser) {
	parser = new(ErrorLogParser)
	parser.init(name, chAlarm, dbFile, dbName, createTable, insertSql)

	parser.skippedErrors = parser.conf.StringList("msg_skip", []string{""})

	go parser.CollectAlarms()

	return
}

func (this *ErrorLogParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DbParser.ParseLine(line)
	if dryRun {
		return
	}

	cls, err := data.Get("class").String()
	if err != nil || cls == "MongoException" {
		// not a error log
		return
	}

	msg, err := data.Get("message").String()
	checkError(err)
	msg = this.normalizeMsg(msg)
	for _, skipped := range this.skippedErrors {
		if strings.Contains(msg, skipped) {
			return
		}
	}
	level, err := data.Get("level").String()
	checkError(err)
	flash, err := data.Get("flash_version_client").String()

	logInfo := extractLogInfo(data)
	this.insert(area, ts, cls, level, msg, flash, logInfo.host)

	return
}

func (this *ErrorLogParser) normalizeMsg(msg string) string {
	r := digitsRegexp.ReplaceAll([]byte(msg), []byte("?"))
	r = tokenRegexp.ReplaceAll(r, []byte("pre cur"))
	return string(r)
}

func (this *ErrorLogParser) CollectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	beepThreshold := this.conf.Int("beep_threshold", 500)
	sleepInterval := time.Duration(this.conf.Int("sleep", 57))
	color := FgRed

	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select count(*) as am, cls, msg from error where ts<=? group by cls, msg order by am desc", tsTo)
		parsersLock.Lock()
		this.echoCheckpoint(color, tsFrom, tsTo, "Error")
		for rows.Next() {
			var cls, msg string
			var amount int64
			err := rows.Scan(&amount, &cls, &msg)
			checkError(err)

			if amount >= int64(beepThreshold) {
				this.beep()
				this.alarmParserPrintf("%8s%20s %s", gofmt.Comma(amount), cls, msg)
			}

			this.colorPrintfLn(color, "%8s%20s %s", gofmt.Comma(amount), cls, msg)
		}
		parsersLock.Unlock()
		rows.Close()

		this.delRecordsBefore(tsTo)
		this.Unlock()

		if this.stopped {
			this.chWait <- true // FIXME only 1 collectAlarm will succeed
			break
		}
	}

}
