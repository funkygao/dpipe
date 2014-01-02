package main

import (
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/locking"
	"os"
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

	globals = engine.DefaultGlobals()
	globals.Debug = options.debug
	globals.Verbose = options.verbose
	globals.DryRun = options.dryrun
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

	setupMaxProcsAndProfiler()

	launchEngine()

	shutdown()
}
