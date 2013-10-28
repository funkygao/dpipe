package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	"github.com/funkygao/gotime"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

// Payment log parser
type PaymentParser struct {
	DbParser
}

// Constructor
func newPaymentParser(name, color string, chAlarm chan<- Alarm, dbFile, dbName, createTable, insertSql string) (parser *PaymentParser) {
	parser = new(PaymentParser)
	parser.init(name, color, chAlarm, dbFile, dbName, createTable, insertSql)

	go parser.CollectAlarms()

	return
}

func (this *PaymentParser) CollectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	sleepInterval := time.Duration(this.conf.Int("sleep", 69))

	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint()
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select sum(amount) as am, type, area, currency from payment where ts<=? group by type, area, currency order by am desc", tsTo)
		parsersLock.Lock()
		this.echoCheckpoint(tsFrom, tsTo, "Revenue")
		totalAmount := float32(0.0)
		for rows.Next() {
			var area, typ, currency string
			var amount int64
			err := rows.Scan(&amount, &typ, &area, &currency)
			checkError(err)

			amount = amount / 100 // 以分为单位，而不是元
			if amount == 0 {
				break
			}

			this.colorPrintfLn("%5s%3s%12s%5s%12.2f", typ, area, gofmt.Comma(amount), currency,
				float32(amount)*CURRENCY_TABLE[currency])

			totalAmount += float32(amount) * CURRENCY_TABLE[currency]
		}
		this.colorPrintfLn("%25s%12.2f", "Total", totalAmount)
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

func (this *PaymentParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DbParser.ParseLine(line)
	if dryRun {
		return
	}

	typ, err := data.Get("type").String()
	if err != nil || typ != "OK" {
		this.colorPrintfLn("Payment %s %s", gotime.TsToString(int(ts)), typ)
		return
	}

	dataBody := data.Get("data")
	uid, err := dataBody.Get("uid").Int()
	if err != nil {
		if verbose {
			logger.Println("null uid", line)
		}

		return
	}
	level, err := dataBody.Get("level").Int()
	amount, err := dataBody.Get("amount").Int()
	checkError(err)
	ref, err := dataBody.Get("trackRef").String()
	item, err := dataBody.Get("item").String()
	checkError(err)
	currency, err := dataBody.Get("currency").String()
	checkError(err)

	logInfo := extractLogInfo(data)
	this.insert(area, logInfo.host, ts, typ, uid, level, amount, ref, item, currency)

	return
}
