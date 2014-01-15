package engine

import (
	"sync"
	"time"
)

type Input interface {
	Plugin

	Run(r InputRunner, h PluginHelper) (err error)
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

	setTickLength(tickLength time.Duration)
	Ticker() (ticker <-chan time.Time)
}

type iRunner struct {
	pRunnerBase

	inChan     chan *PipelinePack
	tickLength time.Duration
	ticker     <-chan time.Time
}

func (this *iRunner) Inject(pack *PipelinePack) {
	if pack.Ident == "" {
		Globals().Fatalf("empty Ident: %s", *pack)
	}

	this.engine.router.InChan() <- pack
}

func (this *iRunner) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *iRunner) Input() Input {
	return this.plugin.(Input)
}

func (this *iRunner) setTickLength(tickLength time.Duration) {
	this.tickLength = tickLength
}

func (this *iRunner) Ticker() (t <-chan time.Time) {
	return this.ticker
}

func (this *iRunner) Start(e *EngineConfig, wg *sync.WaitGroup) error {
	this.engine = e
	this.inChan = e.inputRecycleChan // got the engine PipelinePack pool
	if this.tickLength > 0 {
		this.ticker = time.Tick(this.tickLength)
	}

	go this.runMainloop(e, wg)
	return nil
}

func (this *iRunner) runMainloop(e *EngineConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	globals := Globals()
	for !globals.Stopping {
		if globals.Verbose {
			globals.Printf("[%s] starting", this.name)
		}

		if err := this.Input().Run(this, e); err != nil {
			panic(err)
		}

		if globals.Verbose {
			globals.Printf("[%s] stopped", this.name)
		}

		if globals.Stopping {
			return
		}

		if restart, ok := this.plugin.(Restarting); ok {
			restart.CleanupForRestart()
		}

		if globals.Verbose {
			globals.Printf("Restarting %s\n", this.Name())
		}

		// Re-initialize our plugin with its wrapper
		iw := e.inputWrappers[this.name]
		this.plugin = iw.Create()
	}
}

func NewInputRunner(name string, input Input, pluginCommons *pluginCommons) (r InputRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			name:          name,
			plugin:        input.(Plugin),
			pluginCommons: pluginCommons,
		},
	}
}
