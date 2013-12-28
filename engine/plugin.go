package engine

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
