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
}

func NewMessageRouter() (router *messageRouter) {
	router = new(messageRouter)
	router.inChan = make(chan *PipelinePack, Globals().PluginChanSize)

	return router
}

func (this *messageRouter) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *messageRouter) Start() {
	go this.mainloop()

	Globals().Println("Router started")
}

func (this *messageRouter) mainloop() {
	var (
		globals = Globals()
		ok      = true
		pack    *PipelinePack
		ticker  *time.Ticker
	)

	ticker = time.NewTicker(time.Second * time.Duration(globals.TickerLength))
	defer ticker.Stop()

	for ok {
		runtime.Gosched()

		select {
		case <-ticker.C:
			globals.Printf("processed msg: %d\n", this.processMessageCount)

		case pack, ok = <-this.inChan:
			if !ok {
				globals.Println("Router inChan closed")
				globals.Stopping = true
				break
			}

			atomic.AddInt64(&this.processMessageCount, 1)
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
