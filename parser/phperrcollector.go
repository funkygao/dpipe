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

	matches := phpErrorRegexp.FindAllStringSubmatch(msg, 10000)[0]
	if len(matches) != 7 {
		return
	}

	host, level, file, line, msg := matches[6], matches[2], matches[4], matches[5], matches[3]
	src := filepath.Base(file) + ":" + line

	this.insert(area, ts, msg, level, host, src)

	return
}
