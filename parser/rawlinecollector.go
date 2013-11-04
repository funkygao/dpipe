package parser

import (
	"github.com/funkygao/alser/config"
)

type RawLineCollectorParser struct {
	CollectorParser
}

func newRawLineCollectorParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *RawLineCollectorParser) {
	this = new(RawLineCollectorParser)
	this.init(conf, chUpstream, chDownstream)

	go this.CollectAlarms()

	return
}

func (this *RawLineCollectorParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.AlsParser.ParseLine(line)

	this.colorPrintfLn(this.conf.PrintFormat, msg)
	this.beep()

	return
}
