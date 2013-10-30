package parser

import (
	json "github.com/bitly/go-simplejson"
	"time"
)

type PhpErrorLogParser struct {
	DbParser
}

// Constructor
func newPhpErrorLogParser(name, color string, chAlarm chan<- Alarm, dbFile, dbName, createTable, insertSql string) (parser *PhpErrorLogParser) {
	parser = new(PhpErrorLogParser)
	parser.init(name, color, chAlarm, dbFile, dbName, createTable, insertSql)

	go parser.CollectAlarms()

	return
}

func (this *PhpErrorLogParser) ParseLine(line string) (area string, ts uint64, _ *json.Json) {
	var data string
	area, ts, data = this.splitLine(line)

	matches := phpErrorRegexp.FindAllStringSubmatch(data, 10000)[0]
	host, level, src, msg := matches[6], matches[2], matches[4], matches[3]

	this.insert(area, ts, host, level, src, msg)

	return
}

func (this *PhpErrorLogParser) CollectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	sleepInterval := time.Duration(this.conf.Int("sleep", 13))

	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select count(*) as am, msg, area, host, level from phperror where ts<=? group by msg, area, host order by am desc", tsTo)
		parsersLock.Lock()
		this.echoCheckpoint(tsFrom, tsTo, "PhpError")
		for rows.Next() {
			var area, msg, host, level string
			var amount int64
			err := rows.Scan(&amount, &msg, &area, &host, &level)
			checkError(err)

			this.alarmParserPrintf("%3s %5d%12s%16s %s", area, amount, level, host, msg)

			this.colorPrintfLn("%3s %5d%12s%16s %s", area, amount, level, host, msg)
		}
		this.beep()
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
