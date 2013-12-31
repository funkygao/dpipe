package plugins

import (
	"errors"
	"github.com/funkygao/funpipe/engine"
)

type SkyOutputConfig struct {
	host string
	port int
}

type SkyOutput struct {
}

func (this *SkyOutput) Init(config interface{}) {

}

func (this *SkyOutput) Config() interface{} {
	return SkyOutputConfig{}
}

func (this *SkyOutput) Run(r engine.OutputRunner, c *engine.PipelineConfig) error {

}
