package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

//
type SyslogngInput struct {
	sink string
	addr string

	stopping bool
}

func (this *SyslogngInput) Init(config *conf.Conf) {
	this.sink = config.String("sink", "")
	if this.sink == "" {
		panic("empty sink")
	}
	this.addr = config.String("addr", ":9787")
}

func (this *SyslogngInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	return nil
}

func (this *SyslogngInput) Stop() {
	this.stopping = true
}

func init() {
	engine.RegisterPlugin("SyslogngInput", func() engine.Plugin {
		return new(SyslogngInput)
	})
}
