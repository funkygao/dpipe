/*
ElasticSearch only parser
*/
package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/rule"
	"time"
)

// area,ts,json
type EsParser struct {
	AlsParser
}

func newEsParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *EsParser) {
	this = new(EsParser)
	this.init(conf, chUpstream, chDownstream)

	return
}

func (this *EsParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.AlsParser.ParseLine(line)
	if msg == "" {
		if verbose {
			logger.Printf("got empty msg: %s\n", line)
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

	_, indexJson, err := this.valuesOfJsonKeys(jsonData)
	if err != nil {
		if debug {
			logger.Println(err)
		}

		return
	}

	if this.conf.Indexing {
		indexJson.Set(INDEX_COL_AREA, area)
		indexJson.Set(INDEX_COL_TIMESTAMP, ts)

		date := time.Unix(int64(ts), 0)
		indexer.c <- indexEntry{indexName: this.conf.IndexName, typ: this.conf.Title, date: &date, data: indexJson}
	}

	return
}
