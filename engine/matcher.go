package engine

import (
	conf "github.com/funkygao/jsconf"
	"strings"
)

type MatchRunner struct {
	inChan chan *PipelinePack
	runner PluginRunner
	rule   *conf.Conf
}

func NewMatchRunner(rule *conf.Conf, r PluginRunner) *MatchRunner {
	this := new(MatchRunner)
	this.rule = rule
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

	for pack := range this.inChan {
		if this.match(pack) {
			matchChan <- pack
		} else {
			pack.Recycle()
		}
	}

	close(matchChan)
}

func (this *MatchRunner) match(pack *PipelinePack) bool {
	return true
}
