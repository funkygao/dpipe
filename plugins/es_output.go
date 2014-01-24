// ElasticSerach plugin
package plugins

import (
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib"
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/golib/sortedmap"
	conf "github.com/funkygao/jsconf"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"time"
)

type EsOutput struct {
	flushInterval  time.Duration
	reportInterval time.Duration
	dryRun         bool
	showProgress   bool

	counters      *sortedmap.SortedMap
	bulkMaxConn   int `json:"bulk_max_conn"`
	bulkMaxDocs   int `json:"bulk_max_docs"`
	bulkMaxBuffer int `json:"bulk_max_buffer"` // in Byte
	indexer       *core.BulkIndexer
	stopChan      chan bool
	totalN        int
}

func (this *EsOutput) Init(config *conf.Conf) {
	this.stopChan = make(chan bool)
	api.Domain = config.String("domain", "localhost")
	this.counters = sortedmap.NewSortedMap()
	api.Port = config.String("port", "9200")
	this.reportInterval = time.Duration(config.Int("report_interval", 30)) * time.Second
	this.showProgress = config.Bool("show_progress", true)
	this.flushInterval = time.Duration(config.Int("flush_interval", 30)) * time.Second
	this.bulkMaxConn = config.Int("bulk_max_conn", 20)
	this.bulkMaxDocs = config.Int("bulk_max_docs", 100)
	this.bulkMaxBuffer = config.Int("bulk_max_buffer", 10<<20) // 10 MB
	this.dryRun = config.Bool("dryrun", false)
}

func (this *EsOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		pack         *engine.PipelinePack
		reloadChan   = make(chan interface{})
		ok           = true
		globals      = engine.Globals()
		inChan       = r.InChan()
		reportTicker = time.NewTicker(this.reportInterval)
	)

	this.indexer = core.NewBulkIndexer(this.bulkMaxConn)
	this.indexer.BulkMaxDocs = this.bulkMaxDocs
	this.indexer.BulkMaxBuffer = this.bulkMaxBuffer

	// start the bulk indexer
	this.indexer.Run(this.stopChan)

	defer reportTicker.Stop()

	observer.Subscribe(engine.RELOAD, reloadChan)

LOOP:
	for ok {
		select {
		case <-this.stopChan:
			ok = false

		case <-reportTicker.C:
			this.handlePeriodicalCounters()

		case <-reloadChan:
			// TODO

		case <-time.After(this.flushInterval):
			this.indexer.Flush()

		case pack, ok = <-inChan:
			if !ok {
				break LOOP
			}

			if globals.Debug {
				globals.Println(*pack)
			}

			this.feedEs(h.Project(pack.Project), pack)
			pack.Recycle()
		}
	}

	engine.Globals().Printf("[%s]Total output to ES: %d", r.Name(), this.totalN)

	// before shutdown, flush again
	if globals.Verbose {
		engine.Globals().Println("Waiting for ES flush...")
	}
	this.indexer.Flush()
	if globals.Verbose {
		engine.Globals().Println("ES flushed")
	}

	// let indexer stop
	this.stopChan <- true

	return nil
}

func (this *EsOutput) handlePeriodicalCounters() {
	if !this.showProgress {
		return
	}

	total := 0
	for _, key := range this.counters.SortedKeys() {
		val := this.counters.Get(key)
		if val > 0 {
			total += val
			engine.Globals().Printf("%-50s %12s", key, gofmt.Comma(int64(val)))

			this.counters.Set(key, 0)
		}
	}

	engine.Globals().Printf("%50s %12s", "Sum", gofmt.Comma(int64(total)))
}

func (this *EsOutput) feedEs(project *engine.ConfProject, pack *engine.PipelinePack) {
	if pack.EsType == "" || pack.EsIndex == "" {
		if project.ShowError {
			project.Printf("Empty ES meta: %s plugins:%v",
				*pack, pack.PluginNames())
		}

		this.counters.Inc("_error_", 1)

		return
	}

	this.counters.Inc(pack.EsIndex+":"+pack.EsType, 1)
	this.totalN += 1

	if this.dryRun {
		return
	}

	date := time.Unix(int64(pack.Message.Timestamp), 0)
	data, err := pack.Message.MarshalPayload()
	if err != nil {
		project.Println(err, *pack)
		return
	}
	id, _ := golib.UUID()
	this.indexer.Index(pack.EsIndex, pack.EsType, id, "", &date, data) // ttl empty
}

func init() {
	engine.RegisterPlugin("EsOutput", func() engine.Plugin {
		return new(EsOutput)
	})
}
