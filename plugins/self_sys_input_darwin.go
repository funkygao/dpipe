package plugins

import (
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
	"time"
)

type SelfSysInput struct {
	stopChan chan bool
	interval time.Duration
}

func (this *SelfSysInput) Init(config *conf.Conf) {
	globals := engine.Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *config)
	}

	this.stopChan = make(chan bool)
	this.interval = time.Duration(config.Int("interval", 10))
}

func (this *SelfSysInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {
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

		case <-time.After(this.interval * time.Second):
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
