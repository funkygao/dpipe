package main

import (
	"fmt"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/locking"
	"os"
	"runtime/debug"
	"strings"
)

func init() {
	parseFlags()

	if options.showversion {
		showVersionAndExit()
	}

	if options.lock {
		if locking.InstanceLocked(LOCKFILE) {
			fmt.Fprintf(os.Stderr, "Another instance is running, exit...\n")
			os.Exit(1)
		}
		locking.LockInstance(LOCKFILE)
	}

	// must be after daemonize, or the pid will be parent pid
	logger = newLogger()
}

func main() {
	defer func() {
		cleanup()

		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	setupMaxProcsAndProfiler()

	globals := engine.DefaultGlobals()
	globals.Debug = options.debug
	globals.Verbose = options.verbose
	globals.DryRun = options.dryrun
	globals.Logger = logger

	eng := engine.NewEngineConfig(globals)
	eng.LoadConfigFile(options.config)
	engine.Launch(eng)

	shutdown()
}
