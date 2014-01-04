package engine

import (
	"runtime"
	"sync/atomic"
	"time"
)

type MessageRouter interface {
	InChan() chan *PipelinePack
}

type messageRouter struct {
	inChan              chan *PipelinePack
	processMessageCount int64 // 16 BilionBillion

	addFilterMatcher    chan *Matcher
	removeFilterMatcher chan *Matcher
	removeOutputMatcher chan *Matcher

	fMatchers []*Matcher
	oMatchers []*Matcher
}

func NewMessageRouter() (this *messageRouter) {
	this = new(messageRouter)
	this.inChan = make(chan *PipelinePack, Globals().PluginChanSize)

	this.addFilterMatcher = make(chan *Matcher)
	this.removeFilterMatcher = make(chan *Matcher)
	this.removeOutputMatcher = make(chan *Matcher)

	this.fMatchers = make([]*Matcher, 0, 10)
	this.oMatchers = make([]*Matcher, 0, 10)

	return this
}

func (this *messageRouter) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *messageRouter) AddFilterMatcher() chan *Matcher {
	return this.addFilterMatcher
}

func (this *messageRouter) RemoveFilterMatcher() chan *Matcher {
	return this.removeFilterMatcher
}

func (this *messageRouter) RemoveOutputMatcher() chan *Matcher {
	return this.removeOutputMatcher
}

func (this *messageRouter) Start() {
	go this.runMainloop()

	Globals().Println("Router started")
}

func (this *messageRouter) runMainloop() {
	var (
		globals = Globals()
		ok      = true
		pack    *PipelinePack
		ticker  *time.Ticker
		matcher *Matcher
	)

	ticker = time.NewTicker(time.Second * time.Duration(globals.TickerLength))
	defer ticker.Stop()

	for ok {
		runtime.Gosched()

		select {
		case matcher = <-this.addFilterMatcher:
			if matcher != nil {
				exists := false
				available := -1
				for i, m := range this.fMatchers {
					if m == nil {
						available = i
					}
					if matcher == m {
						exists = true
						break
					}
				}
				if !exists {
					if available != -1 {
						this.fMatchers[available] = matcher
					} else {
						this.fMatchers = append(this.fMatchers, matcher)
					}
				}
			}

		case matcher = <-this.removeFilterMatcher:
			if matcher != nil {
				for i, m := range this.fMatchers {
					if matcher == m {
						//close(m.inChan)
						this.fMatchers[i] = nil
						break
					}
				}
			}

		case matcher = <-this.removeOutputMatcher:
			if matcher != nil {
				for i, m := range this.oMatchers {
					if matcher == m {
						//close(m.inChan)
						this.oMatchers[i] = nil
						break
					}
				}
			}

		case <-ticker.C:
			globals.Printf("processed msg: %d\n", this.processMessageCount)

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Stopping = true
				break
			}

			pack.diagnostics.Reset()

			// messages count auditting
			atomic.AddInt64(&this.processMessageCount, 1)

			for _, matcher = range this.fMatchers {
				if matcher == nil {
					// already removed
					continue
				}

				atomic.AddInt32(&pack.RefCount, 1)
			}
			/*
				for _, matcher = range this.fMatchers {
					if matcher != nil {
						atomic.AddInt32(&pack.RefCount, 1)
						pack.diagnostics.AddStamp(matcher.pluginRunner)
						matcher.inChan <- pack
					}
				}
				for _, matcher = range this.oMatchers {
					if matcher != nil {
						atomic.AddInt32(&pack.RefCount, 1)
						pack.diagnostics.AddStamp(matcher.pluginRunner)
						matcher.inChan <- pack
					}
				}*/
			pack.Recycle()
		}
	}

	/*
		for _, matcher = range this.fMatchers {
			if matcher != nil {
				close(matcher.inChan)
			}
		}
		for _, matcher = range this.oMatchers {
			close(matcher.inChan)
		}*/

	globals.Println("MessageRouter stopped")
}
