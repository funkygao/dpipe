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

// API made available to all plugins providing Heka-wide utility functions.
type PluginHelper interface {

	// Returns an `OutputRunner` for an output plugin registered using the
	// specified name, or ok == false if no output by that name is registered.
	Output(name string) (oRunner OutputRunner, ok bool)

	// Returns an `FilterRunner` for a filter plugin registered using the
	// specified name, or ok == false if no filter by that name is registered.
	Filter(name string) (fRunner FilterRunner, ok bool)

	// Returns the currently running Heka instance's unique PipelineConfig
	// object.
	PipelineConfig() *PipelineConfig

	// Returns a single `DecoderSet` of running decoders for use by any plugin
	// (usually inputs) that wants to decode binary data into a `Message`
	// struct.
	DecoderSet() DecoderSet

	// Expects a loop count value from an existing message (or zero if there's
	// no relevant existing message), returns an initialized `PipelinePack`
	// pointer that can be populated w/ message data and inserted into the
	// Heka pipeline. Returns `nil` if the loop count value provided is
	// greater than the maximum allowed by the Heka instance.
	PipelinePack(msgLoopCount uint) *PipelinePack

	// Returns an input plugin of the given name that provides the
	// StatAccumulator interface, or an error value if such a plugin
	// can't be found.
	StatAccumulator(name string) (statAccum StatAccumulator, err error)
}
