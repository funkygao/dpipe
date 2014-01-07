package plugins

import (
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/golib/stats"
	conf "github.com/funkygao/jsconf"
)

type CardinalityOutput struct {
	counters *stats.CardinalityCounter
	project  string
}

func (this *CardinalityOutput) Init(config *conf.Conf) {
	this.counters = stats.NewCardinalityCounter()
	this.project = config.String("proj", "")
}

func (this *CardinalityOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	var (
		pack      *engine.PipelinePack
		resetChan = make(chan interface{})
		dumpChan  = make(chan interface{})
		ok        = true
		inChan    = r.InChan()
	)

	observer.Subscribe(engine.RELOAD, resetChan)
	observer.Subscribe(engine.SIGUSR1, dumpChan)

	for ok {
		select {
		case <-dumpChan:
			this.dumpCounters(e.Project(this.project))

		case <-resetChan:
			this.resetCounters()

		case pack, ok = <-inChan:
			if !ok {
				break
			}

			if pack.CardinalityKey != "" && pack.CardinalityData != nil {
				this.counters.Add(pack.CardinalityKey, pack.CardinalityData)
			}

			pack.Recycle()
		}
	}

	return nil
}

func (this *CardinalityOutput) dumpCounters(project *engine.ConfProject) {
	for _, key := range this.counters.Categories() {
		project.Printf("%25s[%v] %d\n", key,
			this.counters.StartedAt(key), this.counters.Count(key))
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
