package parser

import (
	json "github.com/bitly/go-simplejson"
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

	key, err := data.Get("key").String()
	if err != nil {
		// not a memcache log
		return
	}

	timeout, err := data.Get("ts").Float64()
	if err != nil {
		timeout, _ = data.Get("timeout").Float64()
	}

	// alarm every occurence
	logInfo := extractLogInfo(data)
	this.colorPrintfLn("%3s%16s %5.2f %40s", area, logInfo.host, timeout, key)
	this.beep()

	return
}
