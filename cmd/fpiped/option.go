package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func parseFlags() {
	flag.BoolVar(&options.verbose, "v", false, "verbose")
	flag.StringVar(&options.configfile, "c", "etc/engine.als.cf", "main config file")
	flag.StringVar(&options.logfile, "l", "", "master log file name")
	flag.StringVar(&options.lockfile, "lockfile", "", "lockfile path")
	flag.BoolVar(&options.showversion, "version", false, "show version")
	flag.BoolVar(&options.debug, "debug", false, "debug mode")
	flag.IntVar(&options.tick, "t", 60*10, "tick interval in seconds")
	flag.BoolVar(&options.dryrun, "dryrun", false, "dry run")
	flag.StringVar(&options.cpuprof, "cpuprof", "", "cpu profiling file")
	flag.StringVar(&options.memprof, "memprof", "", "memory profiling file")
	flag.BoolVar(&options.tail, "tail", true, "tail mode instead of full scan")
	flag.Usage = showUsage

	flag.Parse()

	if options.debug {
		options.verbose = true
	}

	if options.tick <= 0 {
		panic("tick must be possitive")
	}
}

func showUsage() {
	fmt.Fprint(os.Stderr, USAGE)
	flag.PrintDefaults()
}

func showVersionAndExit() {
	fmt.Fprintf(os.Stderr, "%s (build: %s)\n", VERSION, BuildID)
	fmt.Fprintf(os.Stderr, "Built with %s %s for %s/%s\n",
		runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	os.Exit(0)
}

func setupMaxProcsAndProfiler() {
	numCpu := runtime.NumCPU()
	maxProcs := numCpu/2 + 1
	runtime.GOMAXPROCS(numCpu)
	globals.Printf("build[%s] starting with %d/%d CPUs...\n", BuildID, maxProcs, numCpu)

	if options.cpuprof != "" {
		f, err := os.Create(options.cpuprof)
		if err != nil {
			panic(err)
		}

		globals.Printf("CPU profiler %s enabled\n", options.cpuprof)
		pprof.StartCPUProfile(f)
	}

}

func newLogger() *log.Logger {
	var logWriter io.Writer = os.Stdout // default log writer
	var err error
	if options.logfile != "" {
		logWriter, err = os.OpenFile(options.logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
	}

	logOptions := log.Ldate | log.Ltime | log.Lshortfile
	if options.debug {
		logOptions |= log.Lmicroseconds
	}

	prefix := fmt.Sprintf("[%d]", os.Getpid())
	log.SetOutput(logWriter)
	log.SetFlags(logOptions)
	log.SetPrefix(prefix)

	return log.New(logWriter, prefix, logOptions)
}
