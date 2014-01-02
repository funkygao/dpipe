package plugins

import (
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type AlarmOutput struct {
}

func (this *AlarmOutput) Init(config *conf.Conf) {
	globals := engine.Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *config)
	}
}

func (this *AlarmOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	return nil
}

func init() {
	engine.RegisterPlugin("AlarmOutput", func() engine.Plugin {
		return new(AlarmOutput)
	})
}
