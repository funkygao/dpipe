package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type SelfSysInput struct {
	stopChan chan bool
	sink     string
}

func (this *SelfSysInput) Init(config *conf.Conf) {
	this.stopChan = make(chan bool)
	this.sink = config.String("sink", "")
	if this.sink == "" {
		panic("empty sink")
	}
}

func (this *SelfSysInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

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

	if globals.Verbose {
		globals.Printf("[%s] stopped", r.Name())
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
