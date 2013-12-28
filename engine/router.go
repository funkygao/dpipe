package engine

// Public interface exposed by the Heka message router. The message router
// accepts packs on its input channel and then runs them through the
// message_matcher for every running Filter and Output plugin. For plugins
// with a positive match, the pack (and any relevant match group captures)
// will be placed on the plugin's input channel.
type MessageRouter interface {
	// Input channel from which the router gets messages to test against the
	// registered plugin message_matchers.
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
