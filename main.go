package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
)

func init() {
	if _, err := os.Stat(lockfile); err == nil {
		fmt.Fprintf(os.Stderr, "another instance is running, exit\n")
		os.Exit(1)
	}

	file, err := os.Create(lockfile)
	if err != nil {
		panic(err)
	}
	file.Close()

	options = parseFlags()
	options.validate()

	if options.pprof != "" {
		f, err := os.Create(options.pprof)
		if err != nil {
			panic(err)
		}

		logger.Printf("CPU profiler enabled, %s\n", options.pprof)
		pprof.StartCPUProfile(f)
	}

	go trapSignals()
}

func main() {
	defer func() {
		cleanup()

		if e := recover(); e != nil {
			debug.PrintStack()
			fmt.Fprintln(os.Stderr, e)
		}
	}()

	logger = newLogger(options)
	numCpu := runtime.NumCPU()/2 + 1
	runtime.GOMAXPROCS(numCpu)
	logger.Printf("starting with %d CPUs...\n", numCpu)

	jsonConfig := loadConfig(options.config)
	logger.Printf("json config has %d items to guard\n", len(jsonConfig))

	guard(jsonConfig)

	logger.Println("terminated")
}
