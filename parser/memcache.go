// memcache set fail guard

package parser

import (
	json "github.com/bitly/go-simplejson"
)

type MemcacheFailParser struct {
	DefaultParser
}

func (this MemcacheFailParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.DefaultParser.ParseLine(line)

	return
}
