package plugins

import (
	"errors"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/tail"
)

type LogfileInputConfig struct {
	interval int
}

type LogfileInput struct {
	tailConf tail.Config
}

func (this *LogfileInput) Init(config interface{}) {
	conf, ok := config.(*LogfileInputConfig)
	if !ok {
		panic("incompatible config type: LogfileInput")
	}

	return nil
}

func (this *LogfileInput) Config() interface{} {

}

func (this *LogfileInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {

}

func (this *LogfileInput) Stop() {

}
