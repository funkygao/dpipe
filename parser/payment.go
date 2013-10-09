package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gofmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

// Payment log parser
type PaymentParser struct {
	DbParser
}

const PAYMENT_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS payment (
	area CHAR(10),
	host CHAR(20),
	ts INT,
	type VARCHAR(50),
    uid INT(10) NULL,
    level INT,
    amount INT,
    ref VARCHAR(50) NULL,
    item VARCHAR(40),
    currency VARCHAR(20)
);
`

// Constructor
func newPaymentParser(chAlarm chan<- Alarm) *PaymentParser {
	var parser *PaymentParser = new(PaymentParser)
	parser.chAlarm = chAlarm

	parser.createDB(PAYMENT_CREATE_TABLE, "var/payment.sqlite")

	go parser.collectAlarms()

	return parser
}

// 在单位时间内:
// 哪个用户支付的金额超过了阀值
// 哪个地区的支付金额超过了阀值
// 非type=OK的数量超过了阀值
// 某主机上来的支付金额超过了阀值
func (this PaymentParser) collectAlarms() {
	for {
		if this.stopped {
			break
		}

		checkpoint := this.getCheckpoint("select max(ts) from payment")
		if checkpoint == 0 {
			continue
		}

		rows := this.query("select sum(amount) as am, type, area, currency from payment where ts<=? group by type, area, currency order by am desc", checkpoint)
		globalLock.Lock()
		this.logCheckpoint(checkpoint)
		for rows.Next() {
			var area, typ, currency string
			var amount int64
			err := rows.Scan(&amount, &typ, &area, &currency)
			checkError(err)

			amount = amount / 100 // 以分为单位，而不是元
			if amount == 0 {
				break
			}

			logger.Printf("%5s%3s%12s%5s\n", typ, area, gofmt.Comma(amount), currency)
		}
		globalLock.Unlock()
		rows.Close()

		if affected := this.execSql("delete from payment where ts<=?", checkpoint); affected > 0 && verbose {
			logger.Printf("payment %d rows deleted\n", affected)
		}

		time.Sleep(time.Second * 19)
	}
}

func (this PaymentParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DefaultParser.ParseLine(line)
	typ, err := data.Get("type").String()
	if err != nil || typ == "repeat" {
		// not a payment log
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

	insert := "INSERT INTO payment(area, host, ts, type, uid, level, amount, ref, item, currency) VALUES(?,?,?,?,?,?,?,?,?,?)"
	logInfo := extractLogInfo(data)
	this.execSql(insert, area, logInfo.host, ts, typ, uid, level, amount, ref, item, currency)

	return
}

type paymentAlarm struct {
	typ                   string
	uid, level, amount    int
	ref, item, area, host string
}

func (this paymentAlarm) String() string {
	return fmt.Sprintf("%s^%s^%d^%d^%d^%s", "P", this.typ, this.uid, this.level, this.amount, this.ref)
}
