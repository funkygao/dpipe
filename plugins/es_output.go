// ElasticSerach plugin
package plugins

import (
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib"
	"github.com/funkygao/golib/observer"
	conf "github.com/funkygao/jsconf"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"time"
)

type EsOutput struct {
	flushInterval time.Duration
	bulkMaxConn   int `json:"bulk_max_conn"`
	bulkMaxDocs   int `json:"bulk_max_docs"`
	bulkMaxBuffer int `json:"bulk_max_buffer"` // in Byte
	indexer       *core.BulkIndexer
	stopChan      chan bool
}

func (this *EsOutput) Init(config *conf.Conf) {
	this.stopChan = make(chan bool)
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

	// start the bulk indexer
	this.indexer.Run(this.stopChan)

	var (
		pack       *engine.PipelinePack
		reloadChan = make(chan interface{})
		ok         = true
		inChan     = r.InChan()
	)

	observer.Subscribe(engine.RELOAD, reloadChan)

	for ok {
		select {
		case <-this.stopChan:
			ok = false

		case <-reloadChan:
			// TODO

		case <-time.After(this.flushInterval * time.Second):
			this.indexer.Flush()

		case pack, ok = <-inChan:
			if !ok {
				break
			}

			this.feedEs(e.Project(pack.Project), pack)
			pack.Recycle()
		}
	}

	// before shutdown, flush again
	this.indexer.Flush()

	// let indexer stop
	this.stopChan <- true

	return nil
}

func (this *EsOutput) feedEs(project *engine.ConfProject, pack *engine.PipelinePack) {
	index, typ := pack.EsIndex, pack.EsType
	if index == "" || typ == "" {
		project.Printf("invalid pack: %s, %#v, msg: %s\n", pack.Logfile.Base(), *pack,
			pack.Message.RawLine())
		project.Println(index, typ)

		return
	}

	date := time.Unix(int64(pack.Message.Timestamp), 0)
	data, _ := pack.Message.MarshalPayload()
	id, _ := golib.UUID()
	this.indexer.Index(pack.EsIndex, pack.EsType, id, "", &date, data) // ttl empty
}

func init() {
	engine.RegisterPlugin("EsOutput", func() engine.Plugin {
		return new(EsOutput)
	})
}
