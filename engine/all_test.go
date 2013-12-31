package engine

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestDefaultGlobals(t *testing.T) {
	e := NewEngineConfig(nil)
	globals := Globals()
	assert.Equal(t, false, globals.Debug)
	assert.Equal(t, 100, globals.PoolSize)
	assert.Equal(t, 50, globals.PluginChanSize)
	assert.Equal(t, ".", globals.BaseDir)
}

func TestDebugEngineConfig(t *testing.T) {
	globals := DefaultGlobals()
	globals.Debug = true
	e := NewEngineConfig(globals)
	e.LoadConfigFile("../etc/main.cf")
}
