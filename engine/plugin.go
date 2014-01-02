package engine

import (
	"fmt"
)

type Plugin interface {
	Init(config interface{})

	// Name/Typ is required attributes
	Config() interface{}
}

// Indicates a plug-in can handle being restart should it exit before
// heka is shut-down.
type Restarting interface {
	// Is called anytime the plug-in returns during the main Run loop to
	// clean up the plug-in state and determine whether the plugin should
	// be restarted or not.
	CleanupForRestart()
}

func RegisterPlugin(name string, factory func() Plugin) {
	if _, present := availablePlugins[name]; present {
		panic(fmt.Sprintf("plugin: %s cannot register twice", name))
	}

	availablePlugins[name] = factory
}

// A helper object to support delayed plugin creation.
type PluginWrapper struct {
	name          string
	configCreator func() interface{}
	pluginCreator func() Plugin
}

func (this *PluginWrapper) Create() (plugin Plugin) {
	plugin = this.pluginCreator()
	plugin.Init(this.configCreator())
	return
}
