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

	Plugin() Plugin

	// Sets the amount of currently 'leaked' packs that have gone through
	// this plugin. The new value will overwrite prior ones.
	SetLeakCount(count int)

	// Returns the current leak count
	LeakCount() int
}
