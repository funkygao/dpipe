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
	inChan              chan *PipelinePack
	processMessageCount int64 // 16 BilionBillion

	filterMatchers []*MatchRunner
	outputMatchers []*MatchRunner
}

func NewMessageRouter() (this *messageRouter) {
	this = new(messageRouter)
	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)

	this.filterMatchers = make([]*MatchRunner, 0, 10)
	this.outputMatchers = make([]*MatchRunner, 0, 10)

	return this
}

func (this *messageRouter) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *messageRouter) Start() {
	go this.runMainloop()

	Globals().Println("Router started")
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

	ticker = time.NewTicker(time.Second * time.Duration(globals.TickerLength))
	defer ticker.Stop()

	for ok {
		//runtime.Gosched()

		select {
		case <-ticker.C:
			globals.Printf("processed msg: %d\n", this.processMessageCount)

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Stopping = true
				break
			}

			pack.diagnostics.Reset()
			atomic.AddInt64(&this.processMessageCount, 1)

			for _, matcher = range this.filterMatchers {
				pack.diagnostics.AddStamp(matcher.runner)
				pack.IncRef()
				matcher.inChan <- pack
			}
			for _, matcher = range this.outputMatchers {
				pack.diagnostics.AddStamp(matcher.runner)
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
