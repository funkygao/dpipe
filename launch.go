package main

import (
	"fmt"
	"github.com/funkygao/funpipe/engine"
	_ "github.com/funkygao/funpipe/plugins"
	"path/filepath"
	"sync"
	"time"
)

func launchEngine() {
	if options.tick > 0 { // ticker for reporting workers progress
		ticker := time.NewTicker(time.Second * time.Duration(options.tick))
		go runTicker(ticker)
	}

	eng := engine.NewEngineConfig(globals)
	eng.LoadConfigFile(options.config)
	engine.Launch(eng)
}
