package engine

import (
	"github.com/funkygao/golib/gofmt"
	"runtime"
	"sync/atomic"
	"time"
)

type messageRouter struct {
	inChan chan *PipelinePack

	totalInputMsgN     int64
	periodInputMsgN    int32
	totalProcessedMsgN int64 // 16 BilionBillion
	periodProcessMsgN  int32

	removeFilterMatcher chan *Matcher
	removeOutputMatcher chan *Matcher

	filterMatchers []*Matcher
	outputMatchers []*Matcher
}

func NewMessageRouter() (this *messageRouter) {
	this = new(messageRouter)
	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)

	this.removeFilterMatcher = make(chan *Matcher)
	this.removeOutputMatcher = make(chan *Matcher)
	this.filterMatchers = make([]*Matcher, 0, 10)
	this.outputMatchers = make([]*Matcher, 0, 10)

	return this
}

// Dispatch pack from Input to MatchRunners
func (this *messageRouter) Start(routerReady chan<- interface{}) {
	var (
		globals    = Globals()
		ok         = true
		pack       *PipelinePack
		ticker     *time.Ticker
		matcher    *Matcher
		foundMatch bool
	)

	ticker = time.NewTicker(time.Second * time.Duration(globals.TickerLength))
	defer ticker.Stop()

	if globals.Verbose {
		globals.Printf("Router started with ticker %ds\n", globals.TickerLength)
	}

	// tell others to go ahead
	routerReady <- true

DONE:

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
			globals.Printf("Elapsed: %s, Total: %s, speed: %d/s\nInput: %s, speed: %d/s",
				time.Since(globals.StartedAt),
				gofmt.Comma(this.totalProcessedMsgN),
				this.periodProcessMsgN/int32(globals.TickerLength),
				gofmt.Comma(this.totalInputMsgN),
				this.periodInputMsgN/int32(globals.TickerLength))

			this.periodInputMsgN = int32(0)
			this.periodProcessMsgN = int32(0)

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Stopping = true
				break DONE
			}

			atomic.AddInt32(&this.periodProcessMsgN, 1)
			atomic.AddInt64(&this.totalProcessedMsgN, 1)
			if len(pack.diagnostics.Runners()) == 0 {
				// has no runner pack, means Input generated pack
				atomic.AddInt64(&this.totalInputMsgN, 1)
				atomic.AddInt32(&this.periodInputMsgN, 1)
			}

			pack.diagnostics.Reset()
			foundMatch = false

			// got pack from Input, now dispatch
			// for each target, pack will inc ref count
			// and the router will dec ref count only once
			for _, matcher = range this.filterMatchers {
				if matcher != nil && matcher.match(pack) {
					foundMatch = true

					pack.IncRef()
					pack.diagnostics.AddStamp(matcher.runner)
					matcher.InChan() <- pack
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
						matcher.InChan() <- pack
					}
				}
			}

			if !foundMatch {
				globals.Printf("Found no match: %s", *pack)
			}

			// never forget this!
			pack.Recycle()
		}
	}

	if globals.Verbose {
		globals.Println("Starting to shutdown filters and outputs...")
	}

	for _, matcher = range this.filterMatchers {
		if matcher != nil {
			close(matcher.InChan())
		}
	}
	for _, matcher = range this.outputMatchers {
		if matcher != nil {
			close(matcher.InChan())
		}
	}

	if globals.Verbose {
		globals.Println("Filters and outputs chan closed")
	}

}

func (this *messageRouter) removeMatcher(matcher *Matcher, matchers []*Matcher) {
	for idx, m := range matchers {
		if m == matcher {
			close(m.InChan())
			matchers[idx] = nil
			break
		}
	}
}
