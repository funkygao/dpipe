package engine

import (
	"sync"
)

type OutputRunner interface {
	PluginRunner

	InChan() chan *PipelinePack

	Output() Output

	Start(e *EngineConfig, wg *sync.WaitGroup) (err error)

	MatchRunner() *MatchRunner
}

type Output interface {
	Plugin

	Run(r OutputRunner, h PluginHelper) (err error)
}
