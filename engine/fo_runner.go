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

}

func (this *foRunner) Inject(pack *PipelinePack) bool {
	go func() {
		this.c.router.InChan() <- pack
	}()
	return true
}

func (this *foRunner) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *foRunner) Output() Output {
	return this.plugin.(Output)
}

func (this *foRunner) Filter() Filter {
	return this.plugin.(Filter)
}

func (this *foRunner) MatchRunner() *MatchRunner {
	return this.matcher
}
