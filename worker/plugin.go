package worker

type WorkerPlugin interface {
	// Receives rule.ConfWorker
	Init(config interface{}) error
}

func RegisterWorkerPlugin(scheme string, factory func() interface{}) {
	if _, present := AvailablePlugins[scheme]; present {
		panic("plugin: " + scheme + " already registered")
	}

	AvailablePlugins[scheme] = factory
}

// A helper object to support delayed plugin creation.
type WorkerPluginWrapper struct {
	scheme        string
	configCreator func() interface{}
	pluginCreator func() interface{}
}

// Create a new instance of the plugin and return it. Errors are ignored. Call
// CreateWithError if an error is needed.
func (this *WorkerPluginWrapper) Create() (plugin interface{}) {
	plugin, _ = this.CreateWithError()
	return
}

// Create a new instance of the plugin and return it, or nil and appropriate
// error value if this isn't possible.
func (this *WorkerPluginWrapper) CreateWithError() (plugin interface{}, err error) {
	plugin = this.pluginCreator()
	err = plugin.(WorkerPlugin).Init(this.configCreator())
	return
}
