package engine

type Input interface {
	Run(ir InputRunner, c PipelineConfig) (err error)
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
