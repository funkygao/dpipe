package parser

import (
	"fmt"
    json "github.com/bitly/go-simplejson"
	"time"
)

// Memcache set fail log parser
type MemcacheFailParser struct {
    DefaultParser
}

// Constructor
func newMemcacheFailParser(chAlarm chan <- Alarm) *MemcacheFailParser {
	var parser *MemcacheFailParser = new(MemcacheFailParser)
	parser.chAlarm = chAlarm
	return parser
}

func (this MemcacheFailParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
    area, ts, data = this.DefaultParser.ParseLine(line)
    key, err := data.Get("key").String()
    if err != nil {
        // not a memcache log
        return
    }

	logInfo := extractLogInfo(data)
	infoData := make(map[string]string)
	infoData["key"] = key

	alarm := MemcacheAlarm{Area: area, Host: logInfo.host, Info: infoData}
	this.chAlarm <- alarm

    return
}

type MemcacheAlarm struct {
	Area string
	Host string
	Duration time.Duration
	Info map[string]string
	Count int
}

func (this MemcacheAlarm) String() string {
	return fmt.Sprintf("%s^%s^%v^%d^%v", this.Area, this.Host, this.Duration, this.Count, this.Info)
}
