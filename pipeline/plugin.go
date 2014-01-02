package worker

import (
	"fmt"
)

type Plugin interface {
	Init(config interface{}) error
}

// Indicates a plug-in has a specific-to-itself config struct that should be
// passed in to its Init method.
type HasConfigStruct interface {
	ConfigStruct() interface{}
}

// Indicates a plug-in can handle being restart should it exit before
// heka is shut-down.
type Restarting interface {
	// Is called anytime the plug-in returns during the main Run loop to
	// clean up the plug-in state and determine whether the plugin should
	// be restarted or not.
	CleanupForRestart()
}

func RegisterPlugin(name string, factory func() interface{}) {
	if _, present := AvailablePlugins[name]; present {
		panic(fmt.Sprintf("plugin: %s cannot register twice", name))
	}

	AvailablePlugins[name] = factory
}

// A helper object to support delayed plugin creation.
type PluginWrapper struct {
	name          string
	configCreator func() interface{}
	pluginCreator func() interface{}
}

// Create a new instance of the plugin and return it. Errors are ignored. Call
// CreateWithError if an error is needed.
func (this *PluginWrapper) Create() (plugin interface{}) {
	plugin, _ = this.CreateWithError()
	return
}

// Create a new instance of the plugin and return it, or nil and appropriate
// error value if this isn't possible.
func (this *PluginWrapper) CreateWithError() (plugin interface{}, err error) {
	plugin = this.pluginCreator()
	err = plugin.(Plugin).Init(this.configCreator())
	return
}

type Input interface {
	Run(ir InputRunner, h PluginHelper) (err error)
	Stop()
}

type InputRunner interface {
	PluginRunner
	// Input channel from which Inputs can get fresh PipelinePacks, ready to
	// be populated.
	InChan() chan *PipelinePack
	// Associated Input plugin object.
	Input() Input

	SetTickLength(tickLength time.Duration)

	// Returns a ticker channel configured to send ticks at an interval
	// specified by the plugin's ticker_interval config value, if provided.
	Ticker() (ticker <-chan time.Time)

	// Starts Input in a separate goroutine and returns. Should decrement the
	// plugin when the Input stops and the goroutine has completed.
	Start(h PluginHelper, wg *sync.WaitGroup) (err error)
	// Injects PipelinePack into the Heka Router's input channel for delivery
	// to all Filter and Output plugins with corresponding message_matchers.
	Inject(pack *PipelinePack)
}

type Message struct {
	Uuid             []byte
	Timestamp        *int64
	Type             *string
	Logger           *string
	Severity         *int32
	Payload          *string
	EnvVersion       *string
	Pid              *int32
	Hostname         *string
	Fields           []*Field
	XXX_unrecognized []byte
}
