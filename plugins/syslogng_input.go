package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

// Directly recv syslog-ng upstream packets
type SyslogngInput struct {
	ident string
	addr  string

	stopping bool
}

func (this *SyslogngInput) Init(config *conf.Conf) {
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
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
