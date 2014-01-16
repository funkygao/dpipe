package engine

import (
	"fmt"
	conf "github.com/funkygao/jsconf"
	"github.com/gorilla/mux"
	"net/http"
)

// Plugin must have Init method
// Besides, it can have CleanupForRestart
type Plugin interface {
	Init(config *conf.Conf)
}

// If a Plugin implements CleanupForRestart, it will be called on restart
type Restarting interface {
	CleanupForRestart()
}

func RegisterPlugin(name string, factory func() Plugin) {
	if _, present := availablePlugins[name]; present {
		panic(fmt.Sprintf("plugin[%s] cannot register twice", name))
	}

	availablePlugins[name] = factory
}

// A helper object to support delayed plugin creation
type PluginWrapper struct {
	name          string
	configCreator func() *conf.Conf
	pluginCreator func() Plugin
}

func (this *PluginWrapper) Create() (plugin Plugin) {
	plugin = this.pluginCreator()
	plugin.Init(this.configCreator())
	return
}

type PluginHelper interface {
	EngineConfig() *EngineConfig
	PipelinePack(msgLoopCount int) *PipelinePack
	Project(name string) *ConfProject
	HttpApiHandleFunc(path string,
		handlerFunc func(http.ResponseWriter,
			*http.Request, map[string]interface{}) (interface{}, error)) *mux.Route
}
