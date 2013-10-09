package parser

import (
    json "github.com/bitly/go-simplejson"
)

// Errlog parser
type ErrorLogParser struct {
    DbParser
}

// Constructor
func newErrorLogParser(chAlarm chan<- Alarm) *ErrorLogParser {
    var parser *ErrorLogParser = new(ErrorLogParser)
    parser.chAlarm = chAlarm
    return parser
}

func (this ErrorLogParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
    area, ts, data = this.DefaultParser.ParseLine(line)

    return
}
