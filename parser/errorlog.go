package parser

import (
	json "github.com/bitly/go-simplejson"
)

// Errlog parser
type ErrorLogParser struct {
    DefaultParser
}

// Constructor
func newErrorLogParser() *ErrorLogParser {
	parser := new(ErrorLogParser)
	return parser
}

func (this ErrorLogParser) ParseLine(line string, ch chan<- Alarm) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DefaultParser.ParseLine(line, ch)

	return
}
