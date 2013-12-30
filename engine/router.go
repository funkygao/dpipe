package engine

type MessageRouter interface {
	InChan() chan *PipelinePack

	// Channel to facilitate adding a matcher to the router which starts the
	// message flow to the associated filter.
	AddFilterMatcher() chan *MatchRunner

	// Channel to facilitate removing a Filter.  If the matcher exists it will
	// be removed from the router, the matcher channel closed and drained, the
	// filter channel closed and drained, and the filter exited.
	RemoveFilterMatcher() chan *MatchRunner
	// Channel to facilitate removing an Output.  If the matcher exists it will
	// be removed from the router, the matcher channel closed and drained, the
	// output channel closed and drained, and the output exited.
	RemoveOutputMatcher() chan *MatchRunner
}

type messageRouter struct {
	inChan              chan *PipelinePack
	addFilterMatcher    chan *MatchRunner
	removeFilterMatcher chan *MatchRunner
	removeOutputMatcher chan *MatchRunner
	fMatchers           []*MatchRunner
	oMatchers           []*MatchRunner
	processMessageCount int64
}

func NewMessageRouter() (router *messageRouter) {
	router = new(messageRouter)
	router.inChan = make(chan *PipelinePack, Globals().PluginChanSize)
	router.addFilterMatcher = make(chan *MatchRunner, 0)
	router.removeFilterMatcher = make(chan *MatchRunner, 0)
	router.removeOutputMatcher = make(chan *MatchRunner, 0)
	router.fMatchers = make([]*MatchRunner, 0, 10)
	router.oMatchers = make([]*MatchRunner, 0, 10)
	return router
}

func (this *messageRouter) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *messageRouter) Start() {
	go this.mainloop()
	Globals().Logger.Println("Router started...")
}

func (this *messageRouter) mainloop() {
	var matcher *MatchRunner
	var ok = true
	var pack *PipelinePack

	for ok {
		runtime.Gosched()
		select {
		case matcher = <-self.addFilterMatcher:
			exists := false
			available := -1
			for i, m := range self.fMatchers {
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
					self.fMatchers[available] = matcher
				} else {
					self.fMatchers = append(self.fMatchers, matcher)
				}
			}

		case matcher = <-self.removeFilterMatcher:
			for i, m := range self.fMatchers {
				if matcher == m {
					close(m.inChan)
					self.fMatchers[i] = nil
					break
				}
			}

		case matcher = <-self.removeOutputMatcher:
			for i, m := range self.oMatchers {
				if matcher == m {
					close(m.inChan)
					self.oMatchers[i] = nil
					break
				}
			}

		case pack, ok = <-self.inChan:
			if !ok {
				break
			}

			pack.diagnostics.Reset()
			atomic.AddInt64(&self.processMessageCount, 1)
			for _, matcher = range self.fMatchers {
				if matcher != nil {
					atomic.AddInt32(&pack.RefCount, 1)
					pack.diagnostics.AddStamp(matcher.pluginRunner)
					matcher.inChan <- pack
				}
			}
			for _, matcher = range self.oMatchers {
				if matcher != nil {
					atomic.AddInt32(&pack.RefCount, 1)
					pack.diagnostics.AddStamp(matcher.pluginRunner)
					matcher.inChan <- pack
				}
			}
			pack.Recycle()
		}
	}

	for _, matcher = range self.fMatchers {
		if matcher != nil {
			close(matcher.inChan)
		}
	}
	for _, matcher = range self.oMatchers {
		close(matcher.inChan)
	}

	log.Println("MessageRouter stopped.")
}
