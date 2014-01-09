package engine

import (
	//"runtime"
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

func (this *messageRouter) Start() {
	go this.runMainloop()
}

func (this *messageRouter) removeMatcher(matcher *MatchRunner,
	matchers []*MatchRunner) {
	if matcher == nil {
		return
	}

	for idx, m := range matchers {
		if m == matcher {
			close(m.inChan)
			matchers[idx] = nil
			break
		}
	}
}

// Dispatch pack from Input to MatchRunners
func (this *messageRouter) runMainloop() {
	var (
		globals = Globals()
		ok      = true
		pack    *PipelinePack
		ticker  *time.Ticker
		matcher *MatchRunner
	)

	if Globals().Verbose {
		Globals().Printf("Router started with ticker %ds\n", globals.TickerLength)
	}

	ticker = time.NewTicker(time.Second * time.Duration(globals.TickerLength))
	defer ticker.Stop()

	for ok {
		//runtime.Gosched()

		select {
		case matcher = <-this.removeOutputMatcher:
			this.removeMatcher(matcher, this.outputMatchers)

		case matcher = <-this.removeFilterMatcher:
			this.removeMatcher(matcher, this.filterMatchers)

		case <-ticker.C:
			elapsed := time.Since(globals.StartedAt)
			globals.Printf("processed msg: %v, elapsed: %s, speed: %d/s\n",
				this.totalProcessedMsgN, elapsed,
				this.periodProcessMsgN/int64(elapsed.Seconds()+1))
			this.periodProcessMsgN = int32(0)

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Stopping = true
				break
			}

			//pack.diagnostics.Reset()
			atomic.AddInt32(&this.periodProcessMsgN, 1)
			atomic.AddInt64(&this.totalProcessedMsgN, 1)

			for _, matcher = range this.filterMatchers {
				if matcher == nil {
					continue
				}

				//pack.diagnostics.AddStamp(matcher.runner)
				pack.IncRef()
				matcher.inChan <- pack
			}
			for _, matcher = range this.outputMatchers {
				if matcher == nil {
					continue
				}

				//pack.diagnostics.AddStamp(matcher.runner)
				pack.IncRef()
				matcher.inChan <- pack
			}

			// never forget this!
			pack.Recycle()
		}
	}

	for _, matcher = range this.filterMatchers {
		close(matcher.inChan)
	}
	for _, matcher = range this.outputMatchers {
		close(matcher.inChan)
	}

	globals.Println("MessageRouter stopped")
}
