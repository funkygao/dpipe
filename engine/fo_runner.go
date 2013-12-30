package engine

import (
	"sync"
)

type foRunner struct {
	pRunnerBase

	matcher    *MatchRunner
	tickLength time.Duration
	ticker     <-chan time.Time
	inChan     chan *PipelinePack
	c          *PipelineConfig
	retainPack *PipelinePack
	leakCount  int
}

func NewFORunner(name string, plugin Plugin) (r *foRunner) {
	r = &foRunner{
		pRunnerBase: pRunnerBase{
			name:   name,
			plugin: plugin,
		},
	}
	r.inChan = make(chan *PipelinePack, 10)
	return
}

func (this *foRunner) Start(c *PipelineConfig, wg *sync.WaitGroup) error {
	this.c = c


}

func (this *foRunner) execute(c *PipelineConfig, wg *sync.WaitGroup) {
	for 
}
