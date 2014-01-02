package main

import (
	"fmt"
	"github.com/funkygao/funpipe/engine"
	_ "github.com/funkygao/funpipe/plugins"
	"github.com/funkygao/golib/locking"
	"github.com/funkygao/golib/signal"
	"os"
	"runtime/debug"
	"syscall"
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

	signal.IgnoreSignal(syscall.SIGHUP)

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

	if options.tick > 0 { // ticker for reporting workers progress
		ticker := time.NewTicker(time.Second * time.Duration(options.tick))
		defer ticker.Stop()

		go runTicker(ticker)
	}

	eng := engine.NewEngineConfig(globals)
	eng.LoadConfigFile(options.configfile)
	engine.Launch(eng)

	shutdown()
}
