package alsparser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"time"
)

// Memcache set fail log parser
type MemcacheFailParser struct {
	AlsParser
}

// Constructor
func newMemcacheFailParser(name string, chAlarm chan<- Alarm) *MemcacheFailParser {
	var parser *MemcacheFailParser = new(MemcacheFailParser)
	parser.init(name, chAlarm)
	return parser
}

func (this MemcacheFailParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	area, ts, data = this.AlsParser.ParseLine(line)
	if dryRun {
		return
	}

	_, err := data.Get("key").String()
	if err != nil {
		// not a memcache log
		return
	}

	// alarm every occurence
	logInfo := extractLogInfo(data)
	this.alarm(memcacheAlarm{area, logInfo.host, time.Unix(int64(ts), 0)})

	warning := fmt.Sprintf("memcache %3s%16s %s", area, logInfo.host, tsToString(int(ts)))
	this.colorPrintln(FgYellow, warning)
	this.beep()

	return
}

type memcacheAlarm struct {
	area string
	host string
	time time.Time
}

func (this memcacheAlarm) String() string {
	return fmt.Sprintf("%s^%s^%s^%s", "M", this.area, this.host, this.time.Format("01-02-15:04:05"))
}
