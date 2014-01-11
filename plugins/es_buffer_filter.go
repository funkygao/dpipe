package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

// buffering pv, pv latency and the alike statistics before feeding ES
type EsBufferFilter struct {
	sink string
}

func (this *EsBufferFilter) Init(config *conf.Conf) {
	this.sink = config.String("sink", "")
	if this.sink == "" {
		panic("empty sink")
	}
}

func (this *EsBufferFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break
			}

			this.handlePack(pack)
			pack.Recycle()
		}
	}

	return nil
}

func (this *EsBufferFilter) handlePack(pack *engine.PipelinePack) {

}

func init() {
	engine.RegisterPlugin("EsBufferFilter", func() engine.Plugin {
		return new(EsBufferFilter)
	})
}
