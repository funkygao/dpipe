package engine

import (
	"sync"
)

// Base interface for the  plugin runners.
type PluginRunner interface {
	Name() string
	SetName(name string)

	// Underlying plugin object
	Plugin() Plugin

	SetLeakCount(count int)
	LeakCount() int
}

// Base struct for the specialized PluginRunners
type pRunnerBase struct {
	name          string
	plugin        Plugin
	engine        *EngineConfig
	pluginCommons *pluginCommons
	leakCount     int
}

type foRunner struct {
	pRunnerBase

	matcher   *MatchRunner
	inChan    chan *PipelinePack
	engine    *EngineConfig
	leakCount int
}

func (this *pRunnerBase) Name() string {
	return this.name
}

func (this *pRunnerBase) SetName(name string) {
	this.name = name
}

func (this *pRunnerBase) Plugin() Plugin {
	return this.plugin
}

func (this *pRunnerBase) SetLeakCount(count int) {
	this.leakCount = count
}

func (this *pRunnerBase) LeakCount() int {
	return this.leakCount
}

func NewFORunner(name string, plugin Plugin, pluginCommons *pluginCommons) (this *foRunner) {
	this = &foRunner{
		pRunnerBase: pRunnerBase{
			name:          name,
			plugin:        plugin,
			pluginCommons: pluginCommons,
		},
	}

	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)
	return
}

func (this *foRunner) Start(e *EngineConfig, wg *sync.WaitGroup) error {
	this.engine = e

	go this.matcher.Start(this.inChan)
	go this.runMainloop(e, wg)
	return nil
}

func (this *foRunner) runMainloop(e *EngineConfig, wg *sync.WaitGroup) {
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

		// Plugin return from 'Run', they died? or we want to stop?

		if globals.Stopping {
			return
		}

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
	// TODO go func may be too much overhead
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
