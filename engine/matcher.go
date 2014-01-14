package engine

import (
	"strings"
)

type MatchRunner struct {
	inChan  chan *PipelinePack
	runner  PluginRunner
	matches []string
}

func NewMatchRunner(matches []string, r PluginRunner) *MatchRunner {
	this := new(MatchRunner)
	this.matches = matches
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
	if len(this.matches) == 0 {
		matchAll = true
	}

	// the mainloop
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
	for _, match := range this.matches {
		if pack.Ident == match {
			return true
		}
	}

	return false
}

func (this *MatchRunner) Name() string {
	return this.runner.Name()
}
