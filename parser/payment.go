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

		checkpoint := this.getCheckpoint("select max(ts) from payment")

		rows := this.query("select sum(amount) as am, type, area, host from payment where ts<=? group by type, area, host order by am desc", checkpoint)
		for rows.Next() {
			var area, host, typ string
			var am int64
			err := rows.Scan(&am, &typ, &area, &host)
			checkError(err)
			logger.Printf("%5s%3s%16s%8s\n", typ, area, host, gofmt.Comma(am))
		}

		if affected := this.execSql("delete from payment where ts<=?", checkpoint); affected > 0 {
			logger.Printf("payment %d rows deleted\n", affected)
		}

		time.Sleep(time.Second * 5)
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
