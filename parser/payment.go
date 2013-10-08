package parser

import (
	json "github.com/bitly/go-simplejson"
)

// Payment log parser
type PaymentParser struct {
    DefaultParser
}

// Constructor
func newPaymentParser() *PaymentParser {
	parser := new(PaymentParser)
	return parser
}

func (this PaymentParser) ParseLine(line string, ch chan Alarm) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DefaultParser.ParseLine(line)

	return
}
