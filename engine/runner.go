package engine

import (
	"fmt"
	"sync"
)

// Base interface for the  plugin runners.
type PluginRunner interface {
	start(e *EngineConfig, wg *sync.WaitGroup) (err error)

	Name() string

	// Underlying plugin object
	Plugin() Plugin

	setLeakCount(count int)
	LeakCount() int
}

// Filter and Output runner extends PluginRunner
type FilterOutputRunner interface {
	PluginRunner

	InChan() chan *PipelinePack
	Matcher() *Matcher
}

type pRunnerBase struct {
	name          string
	plugin        Plugin
	engine        *EngineConfig
	pluginCommons *pluginCommons
	leakCount     int
}

type foRunner struct {
	pRunnerBase

	matcher   *Matcher
	inChan    chan *PipelinePack
	leakCount int
}

func (this *pRunnerBase) Name() string {
	return this.name
}

func (this *pRunnerBase) Plugin() Plugin {
	return this.plugin
}

func (this *pRunnerBase) setLeakCount(count int) {
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

func (this *foRunner) Matcher() *Matcher {
	return this.matcher
}

func (this *foRunner) Inject(pack *PipelinePack) bool {
	if pack.Ident == "" {
		errmsg := fmt.Sprintf("Plugin %s tries to inject pack with empty ident: %s",
			this.Name(), *pack)
		panic(errmsg)
	}

	this.engine.router.inChan <- pack
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

func (this *foRunner) start(e *EngineConfig, wg *sync.WaitGroup) error {
	this.engine = e

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
			if globals.Verbose {
				globals.Printf("Filter[%s]starting", this.name)
			}

			pluginType = "filter"
			filter.Run(this, e)

			if globals.Verbose {
				globals.Printf("Filter[%s]stopped", this.name)
			}
		} else if output, ok := this.plugin.(Output); ok {
			if globals.Verbose {
				globals.Printf("Output[%s]starting", this.name)
			}

			pluginType = "output"
			output.Run(this, e)

			if globals.Verbose {
				globals.Printf("Output[%s]stopped", this.name)
			}
		} else {
			panic("unkown plugin type")
		}

		// Plugin return from 'Run', they died? or we want to stop?

		if globals.Stopping {
			return
		}

		if restart, ok := this.plugin.(Restarting); ok {
			if !restart.CleanupForRestart() {
				return
			}
		}

		if globals.Verbose {
			globals.Printf("[%s]restarting", this.name)
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
