package parser

import (
	"github.com/funkygao/alser/config"
	"path/filepath"
)

type PhperrorCollectorParser struct {
	CollectorParser
}

func newPhperrorCollectorParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *PhperrorCollectorParser) {
	this = new(PhperrorCollectorParser)
	this.init(conf, chUpstream, chDownstream)

	go this.CollectAlarms()

	return
}

func (this *PhperrorCollectorParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.CollectorParser.ParseLine(line)

	phpErrorMatches := phpErrorRegexp.FindAllStringSubmatch(msg, 10000)
	if len(phpErrorMatches) == 0 {
		// not a php_error msg, give out warning now
		this.colorPrintfLn("%3s %s", area, msg)
		return
	}

	matches := phpErrorMatches[0]
	if len(matches) != 7 {
		return
	}

	level, msg, file, line, host := matches[2], matches[3], matches[4], matches[5], matches[6]
	src := filepath.Base(file) + ":" + line

	this.insert(area, ts, msg, level, host, src)

	return
}
