package main

import (
	"fmt"
	"github.com/funkygao/dpipe/engine"
	_ "github.com/funkygao/dpipe/plugins" // trigger RegisterPlugin(s)
	"github.com/funkygao/golib/locking"
	"os"
	"runtime/debug"
	"time"
)

func init() {
	parseFlags()

	if options.showversion {
		showVersionAndExit()
	}

	if options.lockfile != "" {
		if locking.InstanceLocked(options.lockfile) {
			fmt.Fprintf(os.Stderr, "Another instance is running, exit...\n")
			os.Exit(1)
		}
		locking.LockInstance(options.lockfile)
	}

	globals = engine.DefaultGlobals()
	globals.Debug = options.debug
	globals.Verbose = options.verbose
	globals.DryRun = options.dryrun
	globals.TickerLength = options.tick
	globals.VeryVerbose = options.veryVerbose
	globals.Logger = newLogger()
}

func main() {
	defer func() {
		cleanup()

		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	globals.Println(`
     _       _                _ 
    | |     (_)              | |
  __| |_ __  _ _ __   ___  __| |
 / _  | '_ \| | '_ \ / _ \/ _  |
| (_| | |_) | | |_) |  __/ (_| |
 \__,_| .__/|_| .__/ \___|\__,_|
      | |     | |               
      |_|     |_|`)

	setupProfiler()

	ticker := time.NewTicker(time.Second * time.Duration(options.tick))
	go runWatchdog(ticker)
	defer ticker.Stop()

	eng := engine.NewEngineConfig(globals)
	eng.LoadConfigFile(options.configfile)
	engine.Launch(eng)

	shutdown()
}
