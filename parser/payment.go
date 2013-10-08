package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"time"
)

// Payment log parser
type PaymentParser struct {
    DefaultParser
	stats map[string]int
	start time.Time
}

// Constructor
func newPaymentParser(chAlarm chan <- Alarm) *PaymentParser {
	var parser *PaymentParser = new(PaymentParser)
	parser.chAlarm = chAlarm
	parser.prefix = "P"
	parser.stats = make(map[string]int)
	parser.start = time.Now()
	return parser
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
	//delta := time.Since(this.start)
	this.chAlarm <- paymentAlarm{this.prefix, typ, uid, level, amount, ref, item, area, logInfo.host}

	return
}

type paymentAlarm struct {
	prefix, typ string
	uid, level, amount int
	ref, item, area, host string

}

func (this paymentAlarm) String() string {
	return fmt.Sprintf("%s^%s^%d^%d^%d^%s", this.prefix, this.typ, this.uid, this.level, this.amount, this.ref)
}
