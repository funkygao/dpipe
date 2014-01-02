package engine

import (
	"runtime"
	"sync/atomic"
)

type MessageRouter interface {
	InChan() chan *PipelinePack
	AddRoute(from, to string) error
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

func (this *messageRouter) AddRoute(from, to string) error {
	return nil
}

func (this *messageRouter) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *messageRouter) Start() {
	go this.mainloop()
	Globals().Logger.Println("Router started...")
}

func (this *messageRouter) mainloop() {
	var (
		ok   = true
		pack *PipelinePack
	)

	for ok {
		runtime.Gosched()
		select {
		case pack, ok = <-this.inChan:
			if !ok {
				// inChan closed
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

	//log.Println("MessageRouter stopped.")
}
