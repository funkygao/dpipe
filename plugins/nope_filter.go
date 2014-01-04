package plugins

import (
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type NopeFilter struct {
}

func (this *NopeFilter) Init(config *conf.Conf) {
	engine.Globals().Debugf("%#v\n", *config)
}

func (this *NopeFilter) Run(r engine.OutputRunner, e *engine.EngineConfig) error {

	return nil
}

func init() {
	engine.RegisterPlugin("NopeFilter", func() engine.Plugin {
		return new(NopeFilter)
	})
}
