package worker

var (
	AvailablePlugins = make(map[string]func() interface{})
)
