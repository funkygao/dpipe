package plugins

import (
	"fmt"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

// buffering pv, pv latency and the alike statistics before feeding ES
type EsBufferFilter struct {
	ident    string
	wokers   []*esBufferWorker
	stopChan chan interface{}
}

func (this *EsBufferFilter) Init(config *conf.Conf) {
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}

	this.stopChan = make(chan interface{})
	this.wokers = make([]*esBufferWorker, 0, 10)
	for i := 0; i < len(config.List("workers", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("workers[%d]", i))
		if err != nil {
			panic(err)
		}

		worker := new(esBufferWorker)
		worker.init(section, this.ident, this.stopChan)
		this.wokers = append(this.wokers, worker)
	}
}

func (this *EsBufferFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	var (
		pack    *engine.PipelinePack
		ok      = true
		globals = engine.Globals()
		inChan  = r.InChan()
	)

	for _, worker := range this.wokers {
		go worker.run(r, h)
	}

LOOP:
	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break LOOP
			}

			if globals.Debug {
				globals.Println(*pack)
			}

			this.handlePack(pack)
			pack.Recycle()
		}
	}

	total := 0
	for _, worker := range this.wokers {
		total += worker.summary.N
		worker.flush(r, h)
	}

	// all workers will get notified and stop running
	close(this.stopChan)

	globals.Printf("[%s]Total filtered: %d", r.Name(), total)

	return nil
}

func (this *EsBufferFilter) handlePack(pack *engine.PipelinePack) {
	for _, worker := range this.wokers {
		if worker.camelName == pack.Logfile.CamelCaseName() {
			worker.inject(pack)
		}
	}
}

func init() {
	engine.RegisterPlugin("EsBufferFilter", func() engine.Plugin {
		return new(EsBufferFilter)
	})
}
