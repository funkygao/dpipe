package engine

import (
	"sync"
)

type FilterRunner interface {
	PluginRunner

	InChan() chan *PipelinePack

	Filter() Filter

	Start(e *EngineConfig, wg *sync.WaitGroup) (err error)

	Inject(pack *PipelinePack) bool

	MatchRunner() *MatchRunner
}

type Filter interface {
	Plugin

	Run(r FilterRunner, h PluginHelper) (err error)
}
