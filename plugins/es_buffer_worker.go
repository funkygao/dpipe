package plugins

import (
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
	esType    string
	timestamp uint64

	stopChan chan interface{}
}

func (this *esBufferWorker) init(config *conf.Conf, ident string,
	stopChan chan interface{}) {
	this.camelName = config.String("camel_name", "")
	if this.camelName == "" {
		panic("empty camel_name")
	}

	this.ident = ident
	this.stopChan = stopChan
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

	// prefill the es field name
	switch this.expression {
	case "count":
		this.esField = "count"
	default:
		this.esField = this.expression + "_" + this.fieldName
	}
	this.esType = this.camelName + "_" + this.expression
}

func (this *esBufferWorker) inject(pack *engine.PipelinePack) {
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

func (this *esBufferWorker) flush(r engine.FilterRunner, h engine.PluginHelper) {
	if this.summary.N == 0 {
		return
	}

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

	pack.Message.Timestamp = this.timestamp
	pack.Ident = this.ident
	pack.EsIndex = indexName(h.Project(this.projectName),
		this.indexPattern, time.Unix(int64(this.timestamp), 0))
	pack.EsType = this.esType
	pack.Project = this.projectName
	globals := engine.Globals()
	if globals.Debug {
		globals.Println(*pack)
	}
	r.Inject(pack)

	this.summary.Reset()
}

func (this *esBufferWorker) run(r engine.FilterRunner, h engine.PluginHelper) {
	ever := true
	for ever {
		select {
		case <-time.After(this.interval):
			this.flush(r, h)

		case <-this.stopChan:
			ever = false
		}
	}
}
