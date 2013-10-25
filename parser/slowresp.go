package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"strings"
	"time"
)

type SlowResponseParser struct {
	DbParser
}

// Constructor
func newSlowResponseParser(name string, chAlarm chan<- Alarm, dbFile, createTable, insertSql string) (parser *SlowResponseParser) {
	parser = new(SlowResponseParser)
	parser.init(name, chAlarm, dbFile, createTable, insertSql)

	go parser.collectAlarms()

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
	this.Lock()
	this.insert(area, logInfo.host, uri, ts, t2-t1, t3-t2)
	this.Unlock()

	return
}

func (this *SlowResponseParser) normalizeUri(uri string) string {
	fields := strings.SplitN(uri, "?", 2)
	return fields[0]
}

func (this *SlowResponseParser) collectAlarms() {
	if dryRun {
		return
	}

	color := FgBlue
	sleepInterval := time.Duration(this.conf.Int("sleep", 23))
	beepThreshold := this.conf.Int("beep_threshold", 20)
	for {
		if this.stopped {
			break
		}

		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint("slowresp")
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

			warning := fmt.Sprintf("%8s %40s%3s", gofmt.Comma(amount), uri, area)
			this.colorPrintln(color, warning)
		}
		parsersLock.Unlock()
		rows.Close()

		if affected := this.execSql("delete from slowresp where ts<=?", tsTo); affected > 0 && verbose {
			logger.Printf("slowresp %d rows deleted\n", affected)
		}

		this.Unlock()

	}
}
