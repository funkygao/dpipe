package engine

import (
	"sync"
)

type OutputRunner interface {
	PluginRunner

	// Input channel where Output should be listening for incoming messages.
	InChan() chan *PipelinePack

	// Associated Output plugin instance.
	Output() Output

	Start(e *EngineConfig, wg *sync.WaitGroup) (err error)

	MatchRunner() *MatchRunner
}

type Output interface {
	Plugin

	Run(r OutputRunner, e *EngineConfig) (err error)
}
