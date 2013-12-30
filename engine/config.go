package engine

import (
	"encoding/json"
	"fmt"
	conf "github.com/daviddengcn/go-ljson-conf"
	"github.com/funkygao/funpipe/rule"
	"sync"
	"time"
)

type PipelineConfig struct {
	*conf.Conf

	InputRunners  map[string]InputRunner
	inputWrappers map[string]*PluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*PluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*PluginWrapper

	router *messageRouter

	inputRecycleChan  chan *PipelinePack
	injectRecycleChan chan *PipelinePack
	reportRecycleChan chan *PipelinePack

	hostname  string
	pid       int32
	startedAt time.Time
}

func NewPipelineConfig(globals *GlobalConfigStruct) (this *PipelineConfig) {
	this = new(PipelineConfig)

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
	this.reportRecycleChan = make(chan *PipelinePack, 1)

	this.router = NewMessageRouter()

	this.hostname, _ = os.Hostname()
	this.pid = int32(os.Getpid())
	this.startedAt = time.Now()

	return this
}

func (this *PipelineConfig) LoadConfigFile(fn string) error {
	cf, err := conf.Load(fn)
	if err != nil {
		return err
	}

	this.Conf = cf

	projects := this.List("projects", nil)
	for i := 0; i < len(projects); i++ {
		keyPrefix := fmt.Sprintf("projects[%d].", i)
		projectName := this.String(keyPrefix+"name", "")
	}

	return nil
}

func (this *PipelineConfig) 

func (this *PipelineConfig) ExecuteRuleEngine() {
	var ok bool
	var pluginType string
	for _, w := range ruleEngine.Workers {
		wrapper := new(PluginWrapper)
		wrapper.name = w.Scheme()

		if wrapper.pluginCreator, ok = AvailablePlugins[wrapper.name]; !ok {
			panic(fmt.Sprintf("no plugin[%s] found", wrapper.name))
		}

		plugin := wrapper.pluginCreator()
		var config interface{}
		hasConfigStruct, ok := plugin.(HasConfigStruct)
		if !ok {

		}
		configStruct := hasConfigStruct.ConfigStruct()
		wrapper.configCreator = func() interface{} { return configStruct }
		if err := plugin.(Plugin).Init(w); err != nil {
			panic(err)
		}

		// determine plugin type
		pluginCats := PluginTypeRegex.FindStringSubmatch(w.Typ)
		if len(pluginCats) < 2 {
			panic("Type doesn't contain valid plugin name: " + w.Typ)
		}
		pluginCategory := pluginCats[1]
		if pluginCategory == "Input" {
			this.InputRunners[wrapper.name] = NewInputRunner(wrapper.name, plugin.(Input))
			this.inputWrappers[wrapper.name] = wrapper
		}

		runner := NewFilterOutputRunner(wrapper.name, plugin.(Plugin))
		runner.name = wrapper.name
		runner.matcher = nil

		switch pluginCategory {
		case "Filter":
			this.FilterRunners[runner.name] = runner
		case "Output":
			this.OutputRunners[runner.name] = runner
			this.outputWrappers[runner.name] = wrapper
		}

		wrapper.Create()
	}
}
