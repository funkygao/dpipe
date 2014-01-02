package main

import (
	"github.com/funkygao/funpipe/engine"
	_ "github.com/funkygao/funpipe/plugins"
	"time"
)

func launchEngine() {
	if options.tick > 0 { // ticker for reporting workers progress
		ticker := time.NewTicker(time.Second * time.Duration(options.tick))
		defer ticker.Stop()

		go runTicker(ticker)
	}

	eng := engine.NewEngineConfig(globals)
	eng.LoadConfigFile(options.configfile)
	engine.Launch(eng)
}
