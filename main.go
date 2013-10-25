package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
)

func init() {
	options = parseFlags()

	if options.showversion {
		fmt.Fprintf(os.Stderr, "ALSer %s (build: %s)\n", VERSION, BuildID)
		fmt.Fprintf(os.Stderr, "Built with %s %s for %s/%s\n",
			runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if options.lock {
		if instanceLocked() {
			fmt.Fprintf(os.Stderr, "Another instance is running, exit...\n")
			os.Exit(1)
		}
		lockInstance()
	}

	logger = newLogger(options)

	setupSignals()
}

func main() {
	defer func() {
		cleanup()

		if e := recover(); e != nil {
			debug.PrintStack()
			fmt.Fprintln(os.Stderr, e)
		}
	}()

	jsonConfig := loadJsonConfig(options.config)
	if options.parser != "" && !jsonConfig.hasParser(options.parser) {
		logger.Println("invalid parser:", options.parser)
		logger.Println("valid parsers:", jsonConfig.parsers())
		shutdown()
	}

	numCpu := runtime.NumCPU()
	maxProcs := numCpu/2 + 1
	runtime.GOMAXPROCS(numCpu)
	logger.Printf("starting with %d/%d CPUs...\n", maxProcs, numCpu)

	if options.pprof != "" {
		f, err := os.Create(options.pprof)
		if err != nil {
			panic(err)
		}

		logger.Printf("CPU profiler %s enabled\n", options.pprof)
		pprof.StartCPUProfile(f)
	}

	logger.Printf("%s has %d kinds of logs to guard\n", options.config, len(jsonConfig))

	guard(jsonConfig)

	shutdown()
}
