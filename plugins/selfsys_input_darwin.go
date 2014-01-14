package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type SelfSysInput struct {
	stopChan chan bool
	ident    string
}

func (this *SelfSysInput) Init(config *conf.Conf) {
	this.stopChan = make(chan bool)
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
}

func (this *SelfSysInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	var (
		stopped = false
	)

	for !stopped {
		select {
		case <-this.stopChan:
			stopped = true

		case <-r.Ticker():
			// same effect as sleep
		}
	}

	return nil
}

func (this *SelfSysInput) Stop() {
	close(this.stopChan)
}

func init() {
	engine.RegisterPlugin("SelfSysInput", func() engine.Plugin {
		return new(SelfSysInput)
	})
}
