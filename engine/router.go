package engine

import (
	"github.com/funkygao/golib/gofmt"
	"runtime"
	"sync/atomic"
	"time"
)

type MessageRouter interface {
	InChan() chan *PipelinePack
}

type messageRouter struct {
	inChan chan *PipelinePack

	periodProcessMsgN  int32
	totalProcessedMsgN int64 // 16 BilionBillion

	removeFilterMatcher chan *MatchRunner
	removeOutputMatcher chan *MatchRunner

	filterMatchers []*MatchRunner
	outputMatchers []*MatchRunner
}

func NewMessageRouter() (this *messageRouter) {
	this = new(messageRouter)
	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)

	this.removeFilterMatcher = make(chan *MatchRunner)
	this.removeOutputMatcher = make(chan *MatchRunner)
	this.filterMatchers = make([]*MatchRunner, 0, 10)
	this.outputMatchers = make([]*MatchRunner, 0, 10)

	return this
}

func (this *messageRouter) InChan() chan *PipelinePack {
	return this.inChan
}

// Dispatch pack from Input to MatchRunners
func (this *messageRouter) Start(routerReady chan<- interface{}) {
	var (
		globals    = Globals()
		ok         = true
		pack       *PipelinePack
		ticker     *time.Ticker
		matcher    *MatchRunner
		foundMatch bool
	)

	ticker = time.NewTicker(time.Second * time.Duration(globals.TickerLength))
	defer ticker.Stop()

	if globals.Verbose {
		globals.Printf("Router started with ticker %ds\n", globals.TickerLength)
	}

	// tell others to go ahead
	routerReady <- true

	for ok {
		runtime.Gosched()

		select {
		case matcher = <-this.removeOutputMatcher:
			if matcher != nil {
				this.removeMatcher(matcher, this.outputMatchers)
			}

		case matcher = <-this.removeFilterMatcher:
			if matcher != nil {
				this.removeMatcher(matcher, this.filterMatchers)
			}

		case <-ticker.C:
			globals.Printf("Total msg: %s, elapsed: %s, speed: %d/s",
				gofmt.Comma(this.totalProcessedMsgN),
				time.Since(globals.StartedAt),
				this.periodProcessMsgN/int32(globals.TickerLength))
			this.periodProcessMsgN = int32(0)

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Stopping = true
				break
			}

			pack.diagnostics.Reset()
			atomic.AddInt32(&this.periodProcessMsgN, 1)
			atomic.AddInt64(&this.totalProcessedMsgN, 1)
			foundMatch = false

			// got pack from Input, now dispatch
			// for each target, pack will inc ref count
			// and the router will dec ref count only once
			for _, matcher = range this.filterMatchers {
				if matcher != nil && matcher.match(pack) {
					foundMatch = true

					pack.IncRef()
					pack.diagnostics.AddStamp(matcher.runner)
					matcher.inChan <- pack
				}
			}

			// If we send pack to filterMatchers and then outputMatchers
			// because filter may change pack Ident, and this pack bacuase
			// of shared mem, may match both filterMatcher and now outputMatcher
			// then dup dispatching happens!!!
			//
			// So, we for a give pack, filter sink and output sink is exclusive
			if !foundMatch {
				for _, matcher = range this.outputMatchers {
					if matcher != nil && matcher.match(pack) {
						foundMatch = true

						pack.IncRef()
						pack.diagnostics.AddStamp(matcher.runner)
						matcher.inChan <- pack
					}
				}
			}

			if !foundMatch {
				globals.Printf("Found no match: %s, msg=%s",
					*pack, pack.Message.Payload)
			}

			// never forget this!
			pack.Recycle()
		}
	}

	for _, matcher = range this.filterMatchers {
		if matcher != nil {
			close(matcher.inChan)
		}
	}
	for _, matcher = range this.outputMatchers {
		if matcher != nil {
			close(matcher.inChan)
		}
	}

}

func (this *messageRouter) removeMatcher(matcher *MatchRunner,
	matchers []*MatchRunner) {
	for idx, m := range matchers {
		if m == matcher {
			close(m.inChan)
			matchers[idx] = nil
			break
		}
	}
}
