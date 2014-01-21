package engine

import (
	"github.com/funkygao/golib/gofmt"
	"log"
	"sync/atomic"
	"time"
)

type routerStats struct {
	totalInputMsgN       int64
	periodInputMsgN      int32
	totalInputBytes      int64
	periodInputBytes     int64
	totalProcessedBytes  int64
	totalProcessedMsgN   int64 // 16 BilionBillion
	periodProcessedMsgN  int32
	periodProcessedBytes int64
}

func (this *routerStats) inject(pack *PipelinePack) {
	atomic.AddInt64(&this.totalProcessedBytes, int64(pack.Message.Size()))
	atomic.AddInt64(&this.totalProcessedMsgN, 1)
	atomic.AddInt64(&this.periodProcessedBytes, int64(pack.Message.Size()))
	atomic.AddInt32(&this.periodProcessedMsgN, 1)

	if len(pack.diagnostics.Runners()) == 0 {
		// has no runner pack, means Input generated pack
		atomic.AddInt64(&this.totalInputMsgN, 1)
		atomic.AddInt32(&this.periodInputMsgN, 1)
		atomic.AddInt64(&this.totalInputBytes, int64(pack.Message.Size()))
		atomic.AddInt64(&this.periodInputBytes, int64(pack.Message.Size()))
	}
}

func (this *routerStats) resetPeriodCounters() {
	this.periodProcessedBytes = int64(0)
	this.periodInputBytes = int64(0)
	this.periodInputMsgN = int32(0)
	this.periodProcessedMsgN = int32(0)
}

func (this *routerStats) render(logger *log.Logger, elapsed int) {
	logger.Printf("Total: %10s %10s, speed: %6s/s %10s/s",
		gofmt.Comma(this.totalProcessedMsgN),
		gofmt.ByteSize(this.totalProcessedBytes),
		gofmt.Comma(int64(this.periodProcessedMsgN/int32(elapsed))),
		gofmt.ByteSize(this.periodProcessedBytes/int64(elapsed)))
	logger.Printf("Input: %10s %10s, speed: %6s/s %10s/s",
		gofmt.Comma(int64(this.periodInputMsgN)),
		gofmt.ByteSize(this.periodInputBytes),
		gofmt.Comma(int64(this.periodInputMsgN/int32(elapsed))),
		gofmt.ByteSize(this.periodInputBytes/int64(elapsed)))
}

type messageRouter struct {
	inChan chan *PipelinePack

	stats routerStats

	removeFilterMatcher chan *Matcher
	removeOutputMatcher chan *Matcher

	filterMatchers []*Matcher
	outputMatchers []*Matcher

	closedMatcherChan chan interface{}
}

func NewMessageRouter() (this *messageRouter) {
	this = new(messageRouter)
	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)
	this.stats = routerStats{}
	this.removeFilterMatcher = make(chan *Matcher)
	this.removeOutputMatcher = make(chan *Matcher)
	this.filterMatchers = make([]*Matcher, 0, 10)
	this.outputMatchers = make([]*Matcher, 0, 10)
	this.closedMatcherChan = make(chan interface{})

	return
}

// Dispatch pack from Input to MatchRunners
func (this *messageRouter) Start() {
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

LOOP:
	for ok {
		select {
		case matcher = <-this.removeOutputMatcher:
			go this.removeMatcher(matcher, this.outputMatchers)

		case matcher = <-this.removeFilterMatcher:
			go this.removeMatcher(matcher, this.filterMatchers)

		case <-ticker.C:
			this.stats.render(globals.Logger, globals.TickerLength)
			this.stats.resetPeriodCounters()

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Stopping = true
				break LOOP
			}

			this.stats.inject(pack)

			pack.diagnostics.Reset()
			foundMatch = false

			// If we send pack to filterMatchers and then outputMatchers
			// because filter may change pack Ident, and this pack bacuase
			// of shared mem, may match both filterMatcher and outputMatcher
			// then dup dispatching happens!!!
			//
			// We have to dispatch to Output then Filter to avoid that case
			for _, matcher = range this.outputMatchers {
				// a pack can match several Output
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
				// a pack can match several Filter
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
