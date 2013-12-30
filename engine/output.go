package engine

type OutputRunner interface {
	PluginRunner

	// Input channel where Output should be listening for incoming messages.
	InChan() chan *PipelinePack
	// Associated Output plugin instance.
	Output() Output

	Start(config *PipelineConfig, wg *sync.WaitGroup) (err error)
	// Returns a ticker channel configured to send ticks at an interval
	// specified by the plugin's ticker_interval config value, if provided.
	Ticker() (ticker <-chan time.Time)
	// Retains a pack for future delivery to the plugin when a plugin needs
	// to shut down and wants to retain the pack for the next time its
	// running properly
	RetainPack(pack *PipelinePack)
	// Parsing engine for this Output's message_matcher.
	MatchRunner() *MatchRunner
}

// Heka Output plugin type.
type Output interface {
	Run(or OutputRunner, h PluginHelper) (err error)
}
