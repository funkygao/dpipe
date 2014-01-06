package plugins

import (
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type BatchAlsLogInput struct {
	checkpointFile string
}

func (this *BatchAlsLogInput) Init(config *conf.Conf) {
	this.checkpointFile = config.String("chkpntfile", "")

}

func (this *BatchAlsLogInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {
	return nil
}

func (this *BatchAlsLogInput) Stop() {

}

func init() {
	engine.RegisterPlugin("BatchAlsLogInput", func() engine.Plugin {
		return new(BatchAlsLogInput)
	})
}
