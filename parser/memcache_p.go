package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gotime"
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
	warning := fmt.Sprintf("memcache %3s%16s %s", area, logInfo.host, gotime.TsToString(int(ts)))
	this.colorPrintln(FgYellow, warning)
	this.beep()

	return
}
