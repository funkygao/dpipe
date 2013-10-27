package parser

import (
	"database/sql"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"github.com/funkygao/gotime"
	"regexp"
	"strings"
	"time"
)

// Errlog parser
type ErrorLogParser struct {
	DbParser
}

var (
	digitsRegexp       = regexp.MustCompile(`\d+`)
	tokenRegexp        = regexp.MustCompile(`pre: .*; current: .*`)
	skipErrors         []string
	mongoBeepThreshold int
	errorBeepThreshold int
)

// Constructor
func newErrorLogParser(name string, chAlarm chan<- Alarm, dbFile, createTable, insertSql string) (parser *ErrorLogParser) {
	parser = new(ErrorLogParser)
	parser.init(name, chAlarm, dbFile, createTable, insertSql)

	skipErrors = parser.conf.StringList("error.msg_skip", []string{""})
	mongoBeepThreshold = parser.conf.Int("mongo.beep_threshold", 1)
	errorBeepThreshold = parser.conf.Int("error.beep_threshold", 500)

	go parser.collectAllAlarms()

	return
}

func (this ErrorLogParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DbParser.ParseLine(line)
	if dryRun {
		return
	}

	cls, err := data.Get("class").String()
	if err != nil {
		// not a error log
		return
	}

	level, err := data.Get("level").String()
	checkError(err)
	msg, err := data.Get("message").String()
	checkError(err)
	msg = this.normalizeMsg(msg)
	for _, skipped := range skipErrors {
		if strings.Contains(msg, skipped) {
			return
		}
	}
	flash, err := data.Get("flash_version_client").String()

	logInfo := extractLogInfo(data)
	this.insert(area, ts, cls, level, msg, flash, logInfo.host)

	return
}

func (this ErrorLogParser) normalizeMsg(msg string) string {
	r := digitsRegexp.ReplaceAll([]byte(msg), []byte("?"))
	r = tokenRegexp.ReplaceAll(r, []byte("pre cur"))
	return string(r)
}

func (this *ErrorLogParser) collectAlarms(interval time.Duration, table, query, checkpointWhere, title, color string, onRows func(*sql.Rows, string)) {
	for {
		time.Sleep(time.Second * interval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint(table, checkpointWhere)
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query(query, tsTo)
		parsersLock.Lock()
		this.logCheckpoint(color, tsFrom, tsTo, title)
		for rows.Next() {
			onRows(rows, color)
		}
		parsersLock.Unlock()
		rows.Close()

		deleteSql := "DELETE FROM " + table + " WHERE ts<=? AND " + checkpointWhere
		if affected := this.execSql(deleteSql, tsTo); affected > 0 && verbose {
			logger.Printf("error %d rows deleted\n", affected)
		}
		this.Unlock()

		if this.stopped {
			this.chWait <- true
			break
		}
	}

}

func (this *ErrorLogParser) mongoOnRows(rows *sql.Rows, color string) {
	var msg string
	var amount int64
	err := rows.Scan(&amount, &msg)
	checkError(err)

	if amount >= int64(mongoBeepThreshold) {
		this.beep()
	}

	warning := fmt.Sprintf("%5s %s", gofmt.Comma(amount), msg)
	this.colorPrintln(color, warning)
}

func (this *ErrorLogParser) errorOnRows(rows *sql.Rows, color string) {
	var cls, msg string
	var amount int64
	err := rows.Scan(&amount, &cls, &msg)
	checkError(err)

	if amount >= int64(errorBeepThreshold) {
		this.beep()
	}

	warning := fmt.Sprintf("%8s%20s %s", gofmt.Comma(amount), cls, msg)
	this.colorPrintln(color, warning)
}

func (this ErrorLogParser) collectAllAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	errorSleep := this.conf.Int("error.sleep", 57)
	errorQuery := `select count(*) as am, cls, msg from error where ts<=? and cls != 'MongoException' group by cls, msg order by am desc`
	go this.collectAlarms(time.Duration(errorSleep), "error", errorQuery, `cls != 'MongoException'`, "Error",
		FgRed, this.errorOnRows)

	mongoSleep := this.conf.Int("mongo.sleep", 17)
	mongoQuery := `select count(*) as am, msg from error where ts<=? and cls = 'MongoException' group by msg order by am desc`
	go this.collectAlarms(time.Duration(mongoSleep), "error", mongoQuery, `cls = 'MongoException'`, "MongoException",
		FgCyan, this.mongoOnRows)
}
