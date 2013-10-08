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
    _, err := data.Get("key").String()
    if err != nil {
        // not a memcache log
        return
    }

	// alarm every occurence
	logInfo := extractLogInfo(data)
	alarm := MemcacheAlarm{Area: area, Host: logInfo.host, Time: time.Unix(int64(ts), 0)}
	this.chAlarm <- alarm

    return
}

type MemcacheAlarm struct {
	Area string
	Host string
	Time time.Time
}

func (this MemcacheAlarm) String() string {
	return fmt.Sprintf("%s^%s^%s", this.Area, this.Host, this.Time.Format("01-02-15:04:05"))
}
