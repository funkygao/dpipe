package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"strings"
	"time"
)

type SlowResponseParser struct {
	DbParser
}

// Constructor
func newSlowResponseParser(name string, chAlarm chan<- Alarm, dbFile, dbName, createTable, insertSql string) (parser *SlowResponseParser) {
	parser = new(SlowResponseParser)
	parser.init(name, chAlarm, dbFile, dbName, createTable, insertSql)

	go parser.CollectAlarms()

	return
}

func (this *SlowResponseParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.AlsParser.ParseLine(line)
	if dryRun {
		return
	}

	chkpnt := data.Get("ts")
	t1, _ := chkpnt.Get("t1").Int()
	t2, _ := chkpnt.Get("t2").Int()
	t3, _ := chkpnt.Get("t3").Int()

	uri, _ := data.Get("uri").String()
	uri = this.normalizeUri(uri)

	// alarm every occurence
	logInfo := extractLogInfo(data)
	this.insert(area, logInfo.host, uri, ts, t2-t1, t3-t2)

	return
}

func (this *SlowResponseParser) normalizeUri(uri string) string {
	fields := strings.SplitN(uri, "?", 2)
	return fields[0]
}

func (this *SlowResponseParser) CollectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	color := FgBlue
	sleepInterval := time.Duration(this.conf.Int("sleep", 23))
	beepThreshold := this.conf.Int("beep_threshold", 20)
	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select count(*) as am, uri, area from slowresp where ts<=? group by uri, area order by am desc", tsTo)
		parsersLock.Lock()
		this.logCheckpoint(color, tsFrom, tsTo, "SlowResponse")
		for rows.Next() {
			var area, uri string
			var amount int64
			err := rows.Scan(&amount, &uri, &area)
			checkError(err)

			if int(amount) >= beepThreshold {
				this.beep()
			}

			this.colorPrintfLn(color, "%8s %60s%3s", gofmt.Comma(amount), uri, area)
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
