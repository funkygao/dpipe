package plugins

import (
	"errors"
	"github.com/funkygao/funpipe/engine"
)

type LogfileInputConfig struct {
	interval int
}

type LogfileInput struct {
	tailConf tail.Config
}

func (this *LogfileInput) Init(config interface{}) (err error) {
	conf, ok := config.(*LogfileInputConfig)
	if !ok {
		return errors.New("incompatible config type: LogfileInput")
	}

	return nil
}

func (this *LogfileInput) Run(r engine.InputRunner, c *engine.PipelineConfig) error {

}

func (this *LogfileInput) Stop() {

}

func (this *LogfileInput) ConfigStruct() interface{} {
	return &LogfileInputConfig{}
}
