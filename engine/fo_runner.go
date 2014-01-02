package engine

import (
	"sync"
)

type foRunner struct {
	pRunnerBase

	inChan    chan *PipelinePack
	engine    *EngineConfig
	leakCount int
}

func NewFORunner(name string, plugin Plugin) (r *foRunner) {
	r = &foRunner{
		pRunnerBase: pRunnerBase{
			name:   name,
			plugin: plugin,
		},
	}

	r.inChan = make(chan *PipelinePack, Globals().PluginChanSize)
	return
}

func (this *foRunner) Start(e *EngineConfig, wg *sync.WaitGroup) error {
	this.engine = e
	go this.run(e, wg)
	return nil
}

func (this *foRunner) run(e *EngineConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		pluginType string
		pw         *PluginWrapper
	)

	globals := Globals()
	for !globals.Stopping {
		if filter, ok := this.plugin.(Filter); ok {
			pluginType = "filter"
			filter.Run(this, e)
		} else if output, ok := this.plugin.(Output); ok {
			pluginType = "output"
			output.Run(this, e)
		} else {
			panic("unkown plugin type")
		}

		if globals.Stopping {
			return
		}

		//
		if recon, ok := this.plugin.(Restarting); ok {
			recon.CleanupForRestart()
		}

		// Re-initialize our plugin using its wrapper
		if pluginType == "filter" {
			pw = e.filterWrappers[this.name]
		} else {
			pw = e.outputWrappers[this.name]
		}

		this.plugin = pw.Create()
	}

}

func (this *foRunner) Inject(pack *PipelinePack) bool {
	go func() {
		this.engine.router.InChan() <- pack
	}()
	return true
}

func (this *foRunner) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *foRunner) Output() Output {
	return this.plugin.(Output)
}

func (this *foRunner) Filter() Filter {
	return this.plugin.(Filter)
}
