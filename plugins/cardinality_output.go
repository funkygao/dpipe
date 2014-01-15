package plugins

import (
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/bjtime"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/golib/stats"
	conf "github.com/funkygao/jsconf"
)

type CardinalityOutput struct {
	counters  *stats.CardinalityCounter
	project   string
	intervals map[string]string
}

func (this *CardinalityOutput) Init(config *conf.Conf) {
	this.counters = stats.NewCardinalityCounter()
	this.project = config.String("project", "")
	this.intervals = make(map[string]string)
}

func (this *CardinalityOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		pack      *engine.PipelinePack
		resetChan = make(chan interface{})
		dumpChan  = make(chan interface{})
		ok        = true
		project   = h.Project(this.project)
		inChan    = r.InChan()
	)

	observer.Subscribe(engine.RELOAD, resetChan)
	observer.Subscribe(engine.SIGUSR1, dumpChan)

	for ok {
		select {
		case <-dumpChan:
			this.dumpCounters(project)

		case <-resetChan:
			project.Println("Cardinality all reset")
			this.resetCounters()

		case pack, ok = <-inChan:
			if !ok {
				break
			}

			if pack.CardinalityKey != "" && pack.CardinalityData != nil {
				this.intervals[pack.CardinalityKey] = pack.CardinalityInterval

				this.counters.Add(pack.CardinalityKey, pack.CardinalityData)
			}

			pack.Recycle()
		}
	}

	// before we quit, dump counters
	this.dumpCounters(project)

	return nil
}

func (this *CardinalityOutput) dumpCounters(project *engine.ConfProject) {
	project.Println("Current cardinalities:")
	for _, key := range this.counters.Categories() {
		project.Printf("%15s[%v] %d\n", key,
			bjtime.TsToString(int(this.counters.StartedAt(key).Unix())),
			this.counters.Count(key))
	}
}

func (this *CardinalityOutput) resetCounters() {
	this.counters.ResetAll()
}

func init() {
	engine.RegisterPlugin("CardinalityOutput", func() engine.Plugin {
		return new(CardinalityOutput)
	})
}
