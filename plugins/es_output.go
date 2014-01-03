// ElasticSerach plugin
package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib"
	conf "github.com/funkygao/jsconf"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"strings"
	"time"
)

type EsOutput struct {
	flushInterval time.Duration
	bulkMaxConn   int `json:"bulk_max_conn"`
	bulkMaxDocs   int `json:"bulk_max_docs"`
	bulkMaxBuffer int `json:"bulk_max_buffer"` // in Byte
	indexer       *core.BulkIndexer
	index         string
	stopChan      chan bool
}

func (this *EsOutput) Init(config *conf.Conf) {
	globals := engine.Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *config)
	}

	this.stopChan = make(chan bool)
	this.index = config.String("index", "")
	api.Domain = config.String("domain", "localhost")
	api.Port = config.String("port", "9200")
	this.flushInterval = time.Duration(config.Int("flush_interval", 30))
	this.bulkMaxConn = config.Int("bulk_max_conn", 20)
	this.bulkMaxDocs = config.Int("bulk_max_docs", 100)
	this.bulkMaxBuffer = config.Int("bulk_max_buffer", 10<<20) // 10 MB
}

func (this *EsOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	this.indexer = core.NewBulkIndexer(this.bulkMaxConn)
	this.indexer.BulkMaxDocs = this.bulkMaxDocs
	this.indexer.BulkMaxBuffer = this.bulkMaxBuffer

	// load geoip db
	als.LoadGeoDb(e.String("geodbfile", ""))

	// start the bulk indexer
	this.indexer.Run(this.stopChan)

	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

	for ok {
		select {
		case <-this.stopChan:
			ok = false

		case <-time.After(this.flushInterval * time.Second):
			this.indexer.Flush()

		case pack, ok = <-inChan:
			if !ok {
				// inChan closed, shutdown
				break
			}

			// got pack from engine, pass to ES
			this.feedEs(pack)
			pack.Recycle()
		}

	}

	// before shutdown, flush again
	this.indexer.Flush()

	// let indexer stop
	this.stopChan <- true

	return nil
}

func (this *EsOutput) feedEs(pack *engine.PipelinePack) {
	date := time.Unix(int64(pack.Message.Timestamp), 0)
	data, _ := pack.Message.MarshalPayload()
	id, _ := golib.UUID()

	this.indexer.Index(this.indexName(pack.Project, &date),
		pack.Logfile.BizName(), id, "", &date, data) // ttl empty
}

func (this *EsOutput) indexName(project string, date *time.Time) string {
	const (
		YM           = "@ym"
		INDEX_PREFIX = "fun_"
	)

	if strings.HasSuffix(this.index, YM) {
		prefix := project
		fields := strings.SplitN(this.index, YM, 2)
		if fields[0] != "" {
			// e,g. rs@ym
			prefix = fields[0]
		}

		return fmt.Sprintf("%s%s_%d_%d", INDEX_PREFIX, prefix, date.Year(), int(date.Month()))
	}

	return INDEX_PREFIX + this.index
}

func (this *EsOutput) Stop() {
	close(this.stopChan)
}

func init() {
	engine.RegisterPlugin("EsOutput", func() engine.Plugin {
		return new(EsOutput)
	})
}
