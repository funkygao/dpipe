// ElasticSerach plugin
package plugins

import (
	"github.com/funkygao/als"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib"
	conf "github.com/funkygao/jsconf"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"time"
)

type EsOutputConfig struct {
	FlushInterval int `json:"flush_interval"`
	BulkMaxConn   int `json:"bulk_max_conn"`
	BulkMaxDocs   int `json:"bulk_max_docs"`
	BulkMaxBuffer int `json:"bulk_max_buffer"` // in Byte
}

type EsOutput struct {
	flushInterval time.Duration
	stopChan      chan bool
	indexer       *core.BulkIndexer
}

func (this *EsOutput) Init(config *conf.Conf) {
	globals := engine.Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *config)
	}

	api.Domain = config.String("domain", "localhost")
	api.Port = config.String("port", "9200")

	this.flushInterval = time.Duration(config.Int("flush_interval", 30))

	this.stopChan = make(chan bool)
	this.indexer = core.NewBulkIndexer(config.Int("bulk_max_conn", 20))
	this.indexer.BulkMaxDocs = config.Int("bulk_max_docs", 100)
	this.indexer.BulkMaxBuffer = config.Int("bulk_max_buffer", 10<<20) // 10 MB
}

func (this *EsOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

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

	this.indexer.Index("index", "_type", id, "", &date, data) // ttl empty
}

func (this *EsOutput) Stop() {
	close(this.stopChan)
}

func init() {
	engine.RegisterPlugin("EsOutput", func() engine.Plugin {
		return new(EsOutput)
	})
}
