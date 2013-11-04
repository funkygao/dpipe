package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
)

type RawLineDbParser struct {
	DbParser
}

func newRawLineDbParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *RawLineDbParser) {
	this = new(RawLineDbParser)
	this.init(conf, chUpstream, chDownstream)
	return
}

func (this *RawLineDbParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.AlsParser.ParseLine(line)

	this.colorPrintfLn(this.conf.PrintFormat, args...)
	this.beep()

	return
}
