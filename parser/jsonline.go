package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
)

type JsonLineParser struct {
	AlsParser
}

func newJsonLineParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *JsonLineParser) {
	this = new(LineParser)
	this.init(conf, chUpstream, chDownstream)
	return
}

func (this *JsonLineParser) ParseLine(line string) (area string, ts uint64, msg string) {
	var data *json.Json
	area, ts, data = this.AlsParser.parseJsonLine(line)
	if dryRun {
		return
	}

	args := this.extractValues(data)
	if len(args) == 0 {
		return
	}

	this.colorPrintfLn(this.conf.PrintFormat, area, args...)
	this.beep()

	return
}
