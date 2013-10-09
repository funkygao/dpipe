package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"time"
)

// Errlog parser
type ErrorLogParser struct {
	DbParser
}

const ERRLOG_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS error (
	area CHAR(10),
	host CHAR(20),
	ts INT,
	cls VARCHAR(50),
    level CHAR(20),
    msg VARCHAR(200) NULL,
    flash INT
);
`

// Constructor
func newErrorLogParser(chAlarm chan<- Alarm) *ErrorLogParser {
	var parser *ErrorLogParser = new(ErrorLogParser)
	parser.chAlarm = chAlarm

	parser.createDB(ERRLOG_CREATE_TABLE, "var/error.sqlite")

	go parser.collectAlarms()

	return parser
}

func (this ErrorLogParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DefaultParser.ParseLine(line)
	cls, err := data.Get("class").String()
	if err != nil {
		// not a error log
		return
	}

	level, err := data.Get("level").String()
	checkError(err)
	msg, err := data.Get("message").String()
	checkError(err)
	flash, err := data.Get("flash_version_client").String()

	logInfo := extractLogInfo(data)

	insert := "INSERT INTO error(area, ts, cls, level, msg, flash, host) VALUES(?,?,?,?,?,?,?)"
	this.execSql(insert, area, ts, cls, level, msg, flash, logInfo.host)

	return
}

func (this ErrorLogParser) collectAlarms() {
	for {
		if this.stopped {
			break
		}

		checkpoint := this.getCheckpoint("select max(ts) from error")

		rows := this.query("select count(*) as am, area, cls, msg from error where ts<=? group by area, cls, msg order by am desc", checkpoint)
		globalLock.Lock()
		for rows.Next() {
			var area, cls, msg string
			var amount int64
			err := rows.Scan(&amount, &area, &cls, &msg)
			checkError(err)
			logger.Printf("%7s%5s%15s%s", gofmt.Comma(amount), area, cls, msg)
		}
		globalLock.Unlock()

		if affected := this.execSql("delete from error where ts<=?", checkpoint); affected > 0 && verbose {
			logger.Printf("error %d rows deleted\n", affected)
		}

		time.Sleep(time.Second * 5)
	}

}
