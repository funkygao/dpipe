package plugins

import (
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
)

// Debug only, will print every recved raw msg
type DebugOutput struct {
}

func (this *DebugOutput) Init(config *conf.Conf) {

}

func (this *DebugOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
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

			globals.Printf("got msg: %s\n", pack.Message.RawLine())

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
