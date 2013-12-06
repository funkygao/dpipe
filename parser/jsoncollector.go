package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
	"time"
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
		if verbose {
			logger.Printf("got empty msg: %s\n", line)
		}

		return
	}

	if ts == 0 {
		if verbose {
			logger.Printf("invalid ts: %s\n", line)
		}

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

	args, indexJson, err := this.valuesOfJsonKeys(jsonData)
	if err != nil {
		if debug {
			logger.Println(err)
		}

		return
	}

	if this.conf.Indexing {
		indexJson.Set("area", area)
		indexJson.Set("t", ts)

		date := time.Unix(int64(ts), 0)
		indexer.c <- indexEntry{indexName: this.conf.IndexName, typ: this.conf.Title, date: &date, data: indexJson}
	}

	if this.conf.InstantFormat != "" {
		iargs := append([]interface{}{area}, args...) // 'area' is always 1st col
		this.beep()
		this.colorPrintfLn(this.conf.InstantFormat, iargs...)
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{area, ts}, args...)
	this.insert(args...)

	return
}
