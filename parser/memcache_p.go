package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/gotime"
)

// Memcache set fail log parser
type MemcacheFailParser struct {
	AlsParser
}

// Constructor
func newMemcacheFailParser(name, color string, chAlarm chan<- Alarm) *MemcacheFailParser {
	var parser *MemcacheFailParser = new(MemcacheFailParser)
	parser.init(name, color, chAlarm)
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
	this.colorPrintfLn(FgYellow, "memcache %3s%16s %s", area, logInfo.host, gotime.TsToString(int(ts)))
	this.beep()

	return
}
