package parser

import (
	json "github.com/bitly/go-simplejson"
)

// Payment log parser
type PaymentParser struct {
    DefaultParser
}

// Constructor
func newPaymentParser(chAlarm chan <- Alarm) *PaymentParser {
	var parser *PaymentParser = new(PaymentParser)
	parser.chAlarm = chAlarm
	return parser
}

func (this PaymentParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DefaultParser.ParseLine(line)

	return
}
