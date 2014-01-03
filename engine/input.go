package engine

import (
	"sync"
)

type Input interface {
	Run(r InputRunner, e *EngineConfig) (err error)
	Stop()
}

type InputRunner interface {
	PluginRunner

	// Input channel from which Inputs can get fresh PipelinePacks
	InChan() chan *PipelinePack

	// Associated Input plugin object.
	Input() Input

	Start(e *EngineConfig, wg *sync.WaitGroup) (err error)

	// Injects PipelinePack into the Router's input channel for delivery
	// to all Filter and Output plugins with corresponding matcher.
	Inject(pack *PipelinePack)
}

type iRunner struct {
	pRunnerBase

	inChan chan *PipelinePack
}

func (this *iRunner) Inject(pack *PipelinePack) {
	this.engine.router.InChan() <- pack
}

func (this *iRunner) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *iRunner) Input() Input {
	return this.plugin.(Input)
}

func (this *iRunner) Start(e *EngineConfig, wg *sync.WaitGroup) error {
	this.engine = e
	this.inChan = e.inputRecycleChan // got the engine PipelinePack pool

	go this.runMainloop(e, wg)
	return nil
}

func (this *iRunner) runMainloop(e *EngineConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	globals := Globals()
	for !globals.Stopping {
		if err := this.Input().Run(this, e); err != nil {
			panic(err)
		}

		if globals.Stopping {
			return
		}

		if restart, ok := this.plugin.(Restarting); ok {
			restart.CleanupForRestart()
		}

		// Re-initialize our plugin with its wrapper
		iw := e.inputWrappers[this.name]
		this.plugin = iw.Create()
	}
}

func NewInputRunner(name string, input Input) (r InputRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			name:   name,
			plugin: input.(Plugin),
		},
	}
}
