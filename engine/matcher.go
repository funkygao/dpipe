package engine

import (
	"strings"
)

type MatchRunner struct {
	inChan   chan *PipelinePack
	runner   PluginRunner
	matchers []int
}

func NewMatchRunner(matchers []int, r PluginRunner) *MatchRunner {
	this := new(MatchRunner)
	this.matchers = matchers
	this.runner = r
	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)
	return this
}

// Let my runner start myself
func (this *MatchRunner) Start(matchChan chan *PipelinePack) {
	defer func() {
		if r := recover(); r != nil {
			var (
				err error
				ok  bool
			)
			if err, ok = r.(error); !ok {
				panic(r)
			}
			if !strings.Contains(err.Error(), "send on closed channel") {
				panic(r)
			}
		}
	}()

	globals := Globals()
	if globals.Verbose {
		globals.Printf("MatchRunner for %s started", this.runner.Name())
	}

	matchAll := false
	if len(this.matchers) == 0 {
		matchAll = true
	}

	for pack := range this.inChan {
		if matchAll || this.match(pack) {
			matchChan <- pack
		} else {
			pack.Recycle()
		}
	}

	close(matchChan)
}

func (this *MatchRunner) match(pack *PipelinePack) bool {
	for _, sink := range this.matchers {
		if pack.Message.Sink == sink {
			return true
		}
	}

	return false
}
