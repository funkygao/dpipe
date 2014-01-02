package engine

import (
	"fmt"
	conf "github.com/funkygao/jsconf"
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

	// PipelinePack supply for Input plugins.
	inputRecycleChan chan *PipelinePack

	// PipelinePack supply for Filter plugins (separate pool prevents
	// deadlocks).
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

	if globals.Debug {
		globals.Printf("%#v\n", *globals)
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

func (this *EngineConfig) Project(name string) *ConfProject {
	p, present := this.projects[name]
	if !present {
		return nil
	}

	return &p
}

// For Filter to generate new messages
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

	globals := Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *cf)
	}

	// 'projects' section
	for i := 0; i < len(this.List("projects", nil)); i++ {
		keyPrefix := fmt.Sprintf("projects[%d]", i)
		section, err := this.Section(keyPrefix)
		if err != nil {
			panic(err)
		}
		project := ConfProject{}
		project.FromConfig(section)

		this.projects[project.Name] = project
	}

	// 'plugins' section
	for i := 0; i < len(this.List("plugins", nil)); i++ {
		this.loadPluginSection(fmt.Sprintf("plugins[%d]", i))
	}

	if globals.Debug {
		globals.Printf("%#v\n", *this)
	}
}

func (this *EngineConfig) loadPluginSection(keyPrefix string) {
	var (
		ok      bool
		globals = Globals()
	)

	if globals.Debug {
		globals.Printf("loading section[%s]\n", keyPrefix)
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
	config, err := this.Section(keyPrefix)
	if err != nil {
		panic(err)
	}
	wrapper.configCreator = func() *conf.Conf { return config }

	// plugin Init here
	plugin.Init(config)

	pluginCats := pluginTypeRegex.FindStringSubmatch(pluginType)
	if len(pluginCats) < 2 {
		panic("invalid plugin type: " + pluginType)
	}

	pluginCategory := pluginCats[1]
	switch pluginCategory {
	case "Input":
		this.InputRunners[wrapper.name] = NewInputRunner(wrapper.name, plugin.(Input))
		this.inputWrappers[wrapper.name] = wrapper

	case "Filter":
		runner := NewFORunner(wrapper.name, plugin)
		this.FilterRunners[runner.name] = runner
		this.filterWrappers[runner.name] = wrapper

	case "Output":
		runner := NewFORunner(wrapper.name, plugin)
		this.OutputRunners[runner.name] = runner
		this.outputWrappers[runner.name] = wrapper
	}

}
