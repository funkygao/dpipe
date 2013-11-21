package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
)

// Child of AlsParser with db(sqlite3) features
type JsonCollectorParser struct {
	CollectorParser
}

func newJsonCollectorParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *JsonCollectorParser) {
	this = new(JsonCollectorParser)
	this.init(conf, chUpstream, chDownstream)

	go this.CollectAlarms()

	return
}

func (this *JsonCollectorParser) ParseLine(line string) (area string, ts uint64, msg string) {
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

	args, err := this.valuesOfJsonKeys(jsonData)
	if err != nil {
		return
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{area, ts}, args...)
	this.insert(args...)

	return
}
