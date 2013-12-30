package engine

import (
	"sync"
)

type Input interface {
	Run(r InputRunner, c *PipelineConfig) (err error)
	Stop()
}

type InputRunner interface {
	PluginRunner
	// Input channel from which Inputs can get fresh PipelinePacks, ready to
	// be populated.
	InChan() chan *PipelinePack
	// Associated Input plugin object.
	Input() Input

	SetTickLength(tickLength time.Duration)

	// Returns a ticker channel configured to send ticks at an interval
	// specified by the plugin's ticker_interval config value, if provided.
	Ticker() (ticker <-chan time.Time)

	// Starts Input in a separate goroutine and returns. Should decrement the
	// plugin when the Input stops and the goroutine has completed.
	Start(h PluginHelper, wg *sync.WaitGroup) (err error)

	// Injects PipelinePack into the Heka Router's input channel for delivery
	// to all Filter and Output plugins with corresponding message_matchers.
	Inject(pack *PipelinePack)
}

type iRunner struct {
	pRunnerBase

	input  Input
	inChan *PipelinePack
}

func (this *iRunner) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *iRunner) Input() Input {
	return this.input
}

func (this *iRunner) Start(c *PipelineConfig, wg *sync.WaitGroup) error {
	this.c = c
	this.inChan = c.injectRecycleChan

	return nil
}

func (this *iRunner) Inject(pack *PipelinePack) {
	this.c.router.InChan() <- pack
}

func NewInputRunner(name string, input Input) (r InputRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			name:   name,
			plugin: input.(Plugin),
		},
		input: input,
	}
}
