package parser

import (
    json "github.com/bitly/go-simplejson"
)

// Memcache set fail log parser
type MemcacheFailParser struct {
    DefaultParser
}

// Constructor
func newMemcacheFailParser() *MemcacheFailParser {
	parser := new(MemcacheFailParser)
	return parser
}

func (this MemcacheFailParser) ParseLine(line string, ch chan<- Alarm) (area string, ts uint64, data *json.Json) {
    area, ts, data = this.DefaultParser.ParseLine(line, ch)
    key, err := data.Get("key").String()
    if err != nil {
        // not a memcache log
        return
    }

	info := extractLogInfo(data)
	infoData := make(map[string]string)
	infoData["host"] = info.host
	infoData["key"] = key

	alarm := Alarm{Area: area, Info: infoData}
	ch <- alarm

    return
}
