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

	closedMatcherChan chan interface{}
}

func NewMessageRouter() (this *messageRouter) {
	this = new(messageRouter)
	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)

	this.removeFilterMatcher = make(chan *Matcher)
	this.removeOutputMatcher = make(chan *Matcher)
	this.filterMatchers = make([]*Matcher, 0, 10)
	this.outputMatchers = make([]*Matcher, 0, 10)
	this.closedMatcherChan = make(chan interface{})

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
		globals.Printf("Router started with ticker=%ds\n", globals.TickerLength)
	}

	// tell others to go ahead
	routerReady <- true

LOOP:
	for ok {
		runtime.Gosched()

		select {
		case matcher = <-this.removeOutputMatcher:
			go this.removeMatcher(matcher, this.outputMatchers)

		case matcher = <-this.removeFilterMatcher:
			go this.removeMatcher(matcher, this.filterMatchers)

		case <-ticker.C:
			globals.Printf("Total: %s, speed: %d/s",
				gofmt.Comma(this.totalProcessedMsgN),
				this.periodProcessMsgN/int32(globals.TickerLength))
			globals.Printf("Input: %s, speed: %d/s",
				gofmt.Comma(this.totalInputMsgN),
				this.periodInputMsgN/int32(globals.TickerLength))

			this.periodInputMsgN = int32(0)
			this.periodProcessMsgN = int32(0)

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Stopping = true
				break LOOP
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

			// If we send pack to filterMatchers and then outputMatchers
			// because filter may change pack Ident, and this pack bacuase
			// of shared mem, may match both filterMatcher and outputMatcher
			// then dup dispatching happens!!!
			//
			// We have to dispatch to Output then Filter to avoid that case
			for _, matcher = range this.outputMatchers {
				if matcher.match(pack) {
					foundMatch = true

					pack.IncRef()
					pack.diagnostics.AddStamp(matcher.runner)
					matcher.InChan() <- pack
				}
			}

			// got pack from Input, now dispatch
			// for each target, pack will inc ref count
			// and the router will dec ref count only once
			for _, matcher = range this.filterMatchers {
				if matcher.match(pack) {
					foundMatch = true

					pack.IncRef()
					pack.diagnostics.AddStamp(matcher.runner)
					matcher.InChan() <- pack
				}
			}

			if !foundMatch {
				panic("Found no match: " + pack.String())
			}

			// never forget this!
			pack.Recycle()
		}
	}

	if globals.Verbose {
		globals.Println("Router shutdown.")
	}
}

func (this *messageRouter) removeMatcher(matcher *Matcher, matchers []*Matcher) {
	globals := Globals()
	for _, m := range matchers {
		if m == matcher {
			queuedPacks := len(this.inChan)
			for queuedPacks > 0 {
				if globals.Debug {
					globals.Printf("[router]queued packs: %d", queuedPacks)
				}

				time.Sleep(time.Millisecond * 2)
				queuedPacks = len(this.inChan)
			}

			queuedPacks = len(m.InChan())
			for queuedPacks > 0 {
				if globals.Debug {
					globals.Printf("[%s]queued packs: %d", m.runner.Name(), queuedPacks)
				}

				time.Sleep(time.Millisecond * 2)
				queuedPacks = len(m.InChan())
			}

			if globals.Debug {
				globals.Printf("Close inChan of %s", m.runner.Name())
			}

			close(m.InChan())
			this.closedMatcherChan <- 1

			return
		}
	}
}

func (this *messageRouter) waitForFlush() {
	for i := 0; i < len(this.filterMatchers)+len(this.outputMatchers); i++ {
		<-this.closedMatcherChan
	}
}
