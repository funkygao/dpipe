package engine

import (
	"sync"
)

type FilterRunner interface {
	PluginRunner

	// Input channel on which the Filter should listen for incoming messages
	// to be processed. Closure of the channel signals shutdown to the filter.
	InChan() chan *PipelinePack

	// Associated Filter plugin object.
	Filter() Filter

	Start(e *EngineConfig, wg *sync.WaitGroup) (err error)

	// Hands provided PipelinePack to the Heka Router for delivery to any
	// Filter or Output plugins with a corresponding message_matcher. Returns
	// false and doesn't perform message injection if the message would be
	// caught by the sending Filter's message_matcher.
	Inject(pack *PipelinePack) bool
}

type Filter interface {
	Run(r FilterRunner, e *EngineConfig) (err error)
}
