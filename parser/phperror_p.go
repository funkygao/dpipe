package parser

import (
	json "github.com/bitly/go-simplejson"
	"time"
)

type PhpErrorLogParser struct {
	DbParser
}

// Constructor
func newPhpErrorLogParser(name string, chAlarm chan<- Alarm, dbFile, createTable, insertSql string) (parser *PhpErrorLogParser) {
	parser = new(PhpErrorLogParser)
	parser.init(name, chAlarm, dbFile, createTable, insertSql)

	go parser.collectAlarms()

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

func (this *PhpErrorLogParser) collectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	color := FgYellow
	sleepInterval := time.Duration(this.conf.Int("sleep", 35))
	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint("phperror")
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select count(*) as am, msg, area, host, level from phperror where ts<=? group by msg, area, host order by am desc", tsTo)
		parsersLock.Lock()
		this.logCheckpoint(color, tsFrom, tsTo, "PhpError")
		for rows.Next() {
			var area, msg, host, level string
			var amount int64
			err := rows.Scan(&amount, &msg, &area, &host, &level)
			checkError(err)

			this.colorPrintfLn(color, "%5d%3s%12s%16s %s", amount, area, level, host, msg)
		}
		this.beep()
		parsersLock.Unlock()
		rows.Close()

		this.delRecordsBefore("phperror", tsTo)
		this.Unlock()

		if this.stopped {
			this.chWait <- true
			break
		}

	}

}
