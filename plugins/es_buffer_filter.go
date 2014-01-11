package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"time"
)

type esBufferWorker struct {
	projectName string
	camelName   string
	fieldName   string
	fieldType   string
	expression  string // count, mean, max, min, sum, sd
	interval    time.Duration
}

func (this *esBufferWorker) init(config *conf.Conf) {
	this.camelName = config.String("camel_name", "")
	if this.camelName == "" {
		panic("empty camel_name")
	}

	this.interval = time.Duration(config.Int("interval", 10))
	this.projectName = config.String("project", "")
	this.expression = config.String("expression", "count")
	if this.expression != "count" {
		this.fieldName = config.String("field_name", "")
		if this.fieldName == "" {
			panic("empty field_name")
		}
		this.fieldType = config.String("field_type", "float")
	}
}

func (this *esBufferWorker) run(h engine.PluginHelper) {

}

// buffering pv, pv latency and the alike statistics before feeding ES
type EsBufferFilter struct {
	sink   string
	wokers []*esBufferWorker
}

func (this *EsBufferFilter) Init(config *conf.Conf) {
	this.sink = config.String("sink", "")
	if this.sink == "" {
		panic("empty sink")
	}

	this.wokers = make([]*esBufferWorker, 0, 10)
	for i := 0; i < len(config.List("workers", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("workers[%d]", i))
		if err != nil {
			panic(err)
		}

		worker := new(esBufferWorker)
		worker.init(section)
		this.wokers = append(this.wokers, worker)
	}
}

func (this *EsBufferFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break
			}

			this.handlePack(pack)
			pack.Recycle()
		}
	}

	return nil
}

func (this *EsBufferFilter) handlePack(pack *engine.PipelinePack) {

}

func init() {
	engine.RegisterPlugin("EsBufferFilter", func() engine.Plugin {
		return new(EsBufferFilter)
	})
}
