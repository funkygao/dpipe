// ElasticSerach plugin
package plugins

import (
	"errors"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"time"
)

type EsOutputConfig struct {
	Domain        string
	Port          string
	FlushInterval int `json:"flush_interval"`
	BulkMaxConn   int `json:"bulk_max_conn"`
	BulkMaxDocs   int `json:"bulk_max_docs"`
	BulkMaxBuffer int `json:"bulk_max_buffer"` // in Byte
}

type EsOutput struct {
	*EsOutputConfig

	stopChan chan bool
	indexer  *core.BulkIndexer
}

func (this *EsOutput) Init(config interface{}) {
	conf := config.(*EsOutputConfig)
	this.EsOutputConfig = conf

	api.Domain = conf.Domain
	api.Port = conf.Port

	this.stopChan = make(chan bool)
	this.indexer = core.NewBulkIndexer(conf.BulkMaxConn)
	this.indexer.BulkMaxDocs = conf.BulkMaxDocs
	this.indexer.BulkMaxBuffer = conf.BulkMaxBuffer
}

func (this *EsOutput) Config() interface{} {
	return EsOutputConfig{
		Domain:        "localhost",
		Port:          "9200",
		FlushInterval: 30,
		BulkMaxConn:   20,
		BulkMaxDocs:   100,
		BulkMaxBuffer: 10 << 20, // 10 MB
	}

}

func (this *EsOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
	this.indexer.Run(this.stopChan)

	var (
		pack *engine.PipelinePack
		ok   = true
	)

	for ok {
		select {
		case <-this.stopChan:
			ok = fase

		case <-time.After(this.FlushInterval * time.Second):
			this.indexer.Flush()

		case pack, ok = <-r.InChan():
			if !ok {
				// inChan closed, shutdown
				break
			}

			// got pack from engine, pass to ES
			this.feedEs(pack)
		}

	}

	// before shutdown, flush again
	this.indexer.Flush()

	// let indexer stop
	this.stopChan <- true
}

func (this *EsOutput) feedEs(pack *engine.PipelinePack) {
	date := time.Unix(int64(pack.Message.timestamp), 0)
	data := pack.Message.MarshalPayload()
	id, _ := golib.UUID()
	this.indexer.Index(index, _type, id, "", &date, data) // ttl empty
}

func (this *EsOutput) Stop() {
	close(this.stopChan)
	this.indexer.
}

func init() {
	engine.RegisterPlugin("EsOutput", func() interface{} {
		return new(EsOutput)
	})
}
