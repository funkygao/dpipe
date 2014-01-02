package plugins

import (
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type NopeFilter struct {
}

func (this *NopeFilter) Init(config *conf.Conf) {
	globals := engine.Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *config)
	}
}

func (this *NopeFilter) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	return nil
}

func init() {
	engine.RegisterPlugin("NopeFilter", func() engine.Plugin {
		return new(NopeFilter)
	})
}
