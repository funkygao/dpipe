package engine

import (
	"encoding/json"
	"fmt"
	conf "github.com/funkygao/jsconf"
	"github.com/kr/pretty"
	"os"
	"time"
)

type EngineConfig struct {
	*conf.Conf

	projects map[string]ConfProject

	InputRunners  map[string]InputRunner
	inputWrappers map[string]*PluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*PluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*PluginWrapper

	router *messageRouter

	inputRecycleChan  chan *PipelinePack
	injectRecycleChan chan *PipelinePack

	hostname  string
	pid       int
	startedAt time.Time
}

func NewEngineConfig(globals *GlobalConfigStruct) (this *EngineConfig) {
	this = new(EngineConfig)

	if globals == nil {
		globals = DefaultGlobals()
	}
	Globals = func() *GlobalConfigStruct {
		return globals
	}

	this.InputRunners = make(map[string]InputRunner)
	this.inputWrappers = make(map[string]*PluginWrapper)
	this.FilterRunners = make(map[string]FilterRunner)
	this.filterWrappers = make(map[string]*PluginWrapper)
	this.OutputRunners = make(map[string]OutputRunner)
	this.outputWrappers = make(map[string]*PluginWrapper)

	this.inputRecycleChan = make(chan *PipelinePack, globals.PoolSize)
	this.injectRecycleChan = make(chan *PipelinePack, globals.PoolSize)

	this.projects = make(map[string]ConfProject)

	this.router = NewMessageRouter()

	this.hostname, _ = os.Hostname()
	this.pid = os.Getpid()
	this.startedAt = time.Now()

	return this
}

// For filter to generate new messages
func (this *EngineConfig) PipelinePack(msgLoopCount int) *PipelinePack {
	if msgLoopCount++; msgLoopCount > Globals().MaxMsgLoops {
		return nil
	}

	pack := <-this.injectRecycleChan
	pack.RefCount = 1
	pack.MsgLoopCount = msgLoopCount
	pack.Message.Reset()

	return pack
}

func (this *EngineConfig) LoadConfigFile(fn string) {
	cf, err := conf.Load(fn)
	if err != nil {
		panic(err)
	}

	this.Conf = cf
	if Globals().Debug {
		pretty.Printf("%# v\n", *cf)
	}

	// 'projects' section
	projects := this.List("projects", nil)
	for i := 0; i < len(projects); i++ {
		keyPrefix := fmt.Sprintf("projects[%d].", i)
		projectName := this.String(keyPrefix+"name", "")
		projectLogger := this.String(keyPrefix+"logger", "")
		this.projects[projectName] = ConfProject{Name: projectName, Logger: projectLogger}
	}

	// 'plugins' section
	plugins := this.List("plugins", nil)
	for i := 0; i < len(plugins); i++ {
		this.loadSection(fmt.Sprintf("plugins[%d]", i))
	}
}

func (this *EngineConfig) loadSection(keyPrefix string) {
	var ok bool

	if Globals().Debug {
		pretty.Printf("loading section with key: %s\n", keyPrefix)
	}

	wrapper := new(PluginWrapper)
	wrapper.name = this.String(keyPrefix+".name", "")
	if wrapper.name == "" {
		panic(keyPrefix + " must config 'name' attr")
	}
	pluginType := this.String(keyPrefix+".class", "")
	if pluginType == "" {
		pluginType = wrapper.name
	}

	if wrapper.pluginCreator, ok = availablePlugins[pluginType]; !ok {
		panic("invalid plugin type: " + pluginType)
	}

	plugin := wrapper.pluginCreator()

	var config = plugin.Config()
	// decode config to plugin specific struct
	conf.Decode(keyPrefix, &config)
	wrapper.configCreator = func() interface{} { return config }

	plugin.Init(config)

	pluginCats := pluginTypeRegex.FindStringSubmatch(pluginType)
	if len(pluginCats) < 2 {
		panic("invalid plugin type: " + pluginType)
	}

	pluginCategory := pluginCats[1]
	if pluginCategory == "Input" {
		this.InputRunners[wrapper.name] = NewInputRunner(wrapper.name, plugin.(Input))
		this.inputWrappers[wrapper.name] = wrapper

		return
	}

	runner := NewFORunner(wrapper.name, plugin)
	switch pluginCategory {
	case "Filter":
		this.FilterRunners[runner.name] = runner
		this.filterWrappers[runner.name] = wrapper
	case "Output":
		this.OutputRunners[runner.name] = runner
		this.outputWrappers[runner.name] = wrapper
	}

}
