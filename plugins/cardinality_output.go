package plugins

import (
	"github.com/funkygao/dpipe/engine"
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
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

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
			this.resetCounters()

		case pack, ok = <-inChan:
			if !ok {
				break
			}

			if globals.Debug {
				project.Printf("key:%s interval:%s data:%v", pack.CardinalityKey,
					pack.CardinalityInterval, pack.CardinalityData)
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

	if globals.Verbose {
		globals.Printf("[%s] stopped", r.Name())
	}

	return nil
}

func (this *CardinalityOutput) dumpCounters(project *engine.ConfProject) {
	project.Println("Current cardinalities:")
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
