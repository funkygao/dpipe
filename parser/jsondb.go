package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
)

// Child of AlsParser with db(sqlite3) features
type JsonDbParser struct {
	DbParser
}

func newJsonDbParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *JsonDbParser) {
	this = new(JsonDbParser)
	this.init(conf, chUpstream, chDownstream)

	go this.CollectAlarms()

	return
}

func (this *JsonDbParser) init(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) {
	this.DbParser.init(conf, chUpstream, chDownstream) // super
}

func (this *JsonDbParser) ParseLine(line string) (area string, ts uint64, msg string) {
	var data *json.Json
	area, ts, data = this.AlsParser.parseJsonLine(line)
	if dryRun {
		return
	}

	args, err := this.extractKeyValues(data)
	if err != nil {
		return
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{area, ts}, args...)
	this.insert(args...)

	return
}
