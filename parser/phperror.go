package parser

import (
    json "github.com/bitly/go-simplejson"
)

// Php error log parser
// NOTICE/WARNING/ERROR
type PhpErrorLogParser struct {
    DbParser
}

// Constructor
func newPhpErrorLogParser(chAlarm chan<- Alarm) *PhpErrorLogParser {
    var parser *PhpErrorLogParser = new(PhpErrorLogParser)
    parser.chAlarm = chAlarm
    return parser
}

func (this PhpErrorLogParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
    area, ts, data = this.DefaultParser.ParseLine(line)

    return
}
