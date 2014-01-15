package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/stats"
	conf "github.com/funkygao/jsconf"
	"time"
)

type esBufferWorker struct {
	ident        string
	projectName  string
	camelName    string
	indexPattern string
	fieldName    string
	fieldType    string
	expression   string // count, mean, max, min, sum, sd
	interval     time.Duration

	summary   stats.Summary
	esField   string
	timestamp uint64 // only store latest
}

func (this *esBufferWorker) init(config *conf.Conf, ident string) {
	this.camelName = config.String("camel_name", "")
	if this.camelName == "" {
		panic("empty camel_name")
	}

	this.ident = ident
	this.interval = time.Duration(config.Int("interval", 10)) * time.Second
	this.projectName = config.String("project", "")
	this.indexPattern = config.String("index_pattern", "@ym")
	this.expression = config.String("expression", "count")
	if this.expression != "count" {
		this.fieldName = config.String("field_name", "")
		if this.fieldName == "" {
			panic("empty field_name")
		}
		this.fieldType = config.String("field_type", "float")
	}

	this.summary = stats.Summary{}

	// prefill the es fieldl name
	switch this.expression {
	case "count":
		this.esField = "count"
	default:
		this.esField = this.expression + "_" + this.fieldName
	}
}

func (this esBufferWorker) inject(pack *engine.PipelinePack) {
	this.timestamp = pack.Message.Timestamp

	switch this.expression {
	case "count":
		this.summary.N += 1

	default:
		value, err := pack.Message.FieldValue(this.fieldName, this.fieldType)
		if err != nil {
			globals := engine.Globals()
			if globals.Verbose {
				globals.Printf("[%s]%v", this.camelName, err)
			}

			return
		}

		// add counters
		switch this.fieldType {
		case als.KEY_TYPE_INT, als.KEY_TYPE_MONEY, als.KEY_TYPE_RANGE:
			this.summary.Add(float64(value.(int)))

		case als.KEY_TYPE_FLOAT:
			this.summary.Add(value.(float64))
		}
	}
}

func (this *esBufferWorker) run(r engine.FilterRunner, h engine.PluginHelper) {
	var (
		globals = engine.Globals()
	)

	for !globals.Stopping {
		select {
		case <-time.After(this.interval):
			// generate new pack
			pack := h.PipelinePack(0)

			switch this.expression {
			case "count":
				pack.Message.SetField(this.esField, this.summary.N)
			case "mean":
				pack.Message.SetField(this.esField, this.summary.Mean)
			case "max":
				pack.Message.SetField(this.esField, this.summary.Max)
			case "min":
				pack.Message.SetField(this.esField, this.summary.Min)
			case "sd":
				pack.Message.SetField(this.esField, this.summary.Sd())
			case "sum":
				pack.Message.SetField(this.esField, this.summary.Sum)
			default:
				panic("invalid expression: " + this.expression)
			}

			pack.Ident = this.ident
			pack.EsIndex = indexName(h.Project(this.projectName),
				this.indexPattern, time.Unix(int64(this.timestamp), 0))
			pack.EsType = this.camelName
			pack.Project = this.projectName
			r.Inject(pack)

			this.summary.Reset()
		}

	}

}

// buffering pv, pv latency and the alike statistics before feeding ES
type EsBufferFilter struct {
	ident  string
	wokers []*esBufferWorker
}

func (this *EsBufferFilter) Init(config *conf.Conf) {
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}

	this.wokers = make([]*esBufferWorker, 0, 10)
	for i := 0; i < len(config.List("workers", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("workers[%d]", i))
		if err != nil {
			panic(err)
		}

		worker := new(esBufferWorker)
		worker.init(section, this.ident)
		this.wokers = append(this.wokers, worker)
	}
}

func (this *EsBufferFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

	for _, worker := range this.wokers {
		go worker.run(r, h)
	}

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
