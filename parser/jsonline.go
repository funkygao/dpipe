package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
)

// area,ts,{}
type JsonLineParser struct {
	AlsParser
}

func newJsonLineParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *JsonLineParser) {
	this = new(JsonLineParser)
	this.init(conf, chUpstream, chDownstream)
	return
}

func (this *JsonLineParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.AlsParser.ParseLine(line)
	if msg == "" {
		return
	}

	var (
		jsonData *json.Json
		err      error
	)
	jsonData, err = this.msgToJson(msg)
	if err != nil {
		logger.Printf("[%s]invalid json msg: %s", this.id(), msg)
		return
	}

	if dryRun {
		return
	}

	args, err := this.valuesOfKeys(jsonData)
	if err != nil {
		return
	}

	args = append([]interface{}{area}, args...)
	this.colorPrintfLn(this.conf.PrintFormat, args...)
	this.beep()

	return
}
