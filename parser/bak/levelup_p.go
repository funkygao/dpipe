package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"time"
)

type LevelUpParser struct {
	DbParser
}

// Constructor
func newLevelUpParser(name, color string, chAlarm chan<- Alarm, dbFile, dbName, createTable, insertSql string) (parser *LevelUpParser) {
	parser = new(LevelUpParser)
	parser.init(name, color, chAlarm, dbFile, dbName, createTable, insertSql)

	go parser.CollectAlarms()

	return
}

func (this *LevelUpParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.AlsParser.ParseLine(line)
	if dryRun {
		return
	}

	from, err := data.Get("from").Int()
	if err != nil {
		// not a valid levelup log
		return
	}

	this.insert(area, ts, from)

	return
}

func (this *LevelUpParser) CollectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	sleepInterval := time.Duration(this.conf.Int("sleep", 95))
	amountThreshold := this.conf.Int("amount_threshold", 10)

	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select count(*) as am, fromlevel from levelup where ts<=? group by fromlevel order by am desc", tsTo)
		parsersLock.Lock()
		this.echoCheckpoint(tsFrom, tsTo, "LevelUp")
		for rows.Next() {
			var fromlevel int
			var amount int
			err := rows.Scan(&amount, &fromlevel)
			checkError(err)

			if amount < amountThreshold {
				break
			}

			this.colorPrintfLn("%8s %3d ->%3d", gofmt.Comma(int64(amount)), fromlevel, fromlevel+1)
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