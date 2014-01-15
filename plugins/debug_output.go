package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

// Debug only, will print every recved raw msg
type DebugOutput struct {
	blackhole bool
}

func (this *DebugOutput) Init(config *conf.Conf) {
	this.blackhole = config.Bool("blackhole", false)
}

func (this *DebugOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		globals = engine.Globals()
		pack    *engine.PipelinePack
		ok      = true
		inChan  = r.InChan()
	)

	for ok && !globals.Stopping {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break
			}

			if !this.blackhole {
				globals.Println(*pack)
			}

			pack.Recycle()
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("DebugOutput", func() engine.Plugin {
		return new(DebugOutput)
	})
}
