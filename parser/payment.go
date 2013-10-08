package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
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
    item VARCHAR(40)
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

		rows := this.query("select area,host,type, uid, ts from payment")
		for rows.Next() {
			var area, host, typ string
			var uid, ts int
			err := rows.Scan(&area, &host, &typ, &uid, &ts)
			checkError(err)
			logger.Println("haha", area, host, typ, uid, ts)
		}

		time.Sleep(time.Second * 5)
	}

	//delta := time.Since(this.start)
	//this.chAlarm <- paymentAlarm{typ, uid, level, amount, ref, item, area, logInfo.host}

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
	checkError(err)
	level, err := dataBody.Get("level").Int()
	checkError(err)
	amount, err := dataBody.Get("amount").Int()
	checkError(err)
	ref, err := dataBody.Get("trackRef").String()
	checkError(err)
	item, err := dataBody.Get("item").String()
	checkError(err)

	logInfo := extractLogInfo(data)

	insert := "INSERT INTO payment(area, host, ts, type, uid, level, amount, ref, item) VALUES(?,?,?,?,?,?,?,?,?)"
	this.execSql(insert, area, logInfo.host, ts, typ, uid, level, amount, ref, item)

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
