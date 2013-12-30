package engine

// Heka PluginRunner interface for Filter type plugins.
type FilterRunner interface {
	PluginRunner

	// Input channel on which the Filter should listen for incoming messages
	// to be processed. Closure of the channel signals shutdown to the filter.
	InChan() chan *PipelinePack
	// Associated Filter plugin object.
	Filter() Filter

	Start(c *PipelineConfig, wg *sync.WaitGroup) (err error)

	Ticker() (ticker <-chan time.Time)
	// Hands provided PipelinePack to the Heka Router for delivery to any
	// Filter or Output plugins with a corresponding message_matcher. Returns
	// false and doesn't perform message injection if the message would be
	// caught by the sending Filter's message_matcher.
	Inject(pack *PipelinePack) bool
	// Parsing engine for this Filter's message_matcher.
	MatchRunner() *MatchRunner
	// Retains a pack for future delivery to the plugin when a plugin needs
	// to shut down and wants to retain the pack for the next time its
	// running properly
	RetainPack(pack *PipelinePack)
}

type Filter interface {
	Run(r FilterRunner, c *PipelineConfig) (err error)
}
