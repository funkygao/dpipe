package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

// pv, pv latency and the alike statistics before feeding ES
type StatsFilter struct {
	blackhole bool
}

func (this *StatsFilter) Init(config *conf.Conf) {
	this.blackhole = config.Bool("blackhole", false)
}

func (this *StatsFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

	for ok && !globals.Stopping {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break
			}

			if !this.blackhole {
				globals.Printf("got msg: %s\n", pack.Message.RawLine())
			}

			pack.Recycle()
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("StatsFilter", func() engine.Plugin {
		return new(StatsFilter)
	})
}
