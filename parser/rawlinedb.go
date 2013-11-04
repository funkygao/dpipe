package parser

import (
	"github.com/funkygao/alser/config"
)

type RawLineDbParser struct {
	DbParser
}

func newRawLineDbParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *RawLineDbParser) {
	this = new(RawLineDbParser)
	this.init(conf, chUpstream, chDownstream)

	go this.CollectAlarms()

	return
}

func (this *RawLineDbParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.AlsParser.ParseLine(line)

	this.colorPrintfLn(this.conf.PrintFormat, msg)
	this.beep()

	return
}
