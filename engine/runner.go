/*
               PluginRunner
          ---------------------------------
         |             |                   |
   InputRunner     FilterRunner        OutputRunner
*/
package engine

// Base interface for the  plugin runners.
type PluginRunner interface {
	Name() string
	SetName(name string)

	// Underlying plugin object
	Plugin() Plugin

	// Sets the amount of currently 'leaked' packs that have gone through
	// this plugin. The new value will overwrite prior ones.
	SetLeakCount(count int)
	LeakCount() int
}

// Base struct for the specialized PluginRunners
type pRunnerBase struct {
	name      string
	plugin    Plugin
	engine    *EngineConfig
	leakCount int
}

func (this *pRunnerBase) Name() string {
	return this.name
}

func (this *pRunnerBase) SetName(name string) {
	this.name = name
}

func (this *pRunnerBase) Plugin() Plugin {
	return this.plugin
}

func (this *pRunnerBase) SetLeakCount(count int) {
	this.leakCount = count
}

func (this *pRunnerBase) LeakCount() int {
	return this.leakCount
}
