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

func (this MemcacheFailParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
    area, ts, data = this.DefaultParser.ParseLine(line)
    key, err := data.Get("key").String()
    if err != nil {
        // not a memcache log
        return
    }
    println(key)

    return
}
