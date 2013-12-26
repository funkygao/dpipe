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
	flag.StringVar(&options.config, "c", "etc/alser.cf", "config json file")
	flag.StringVar(&options.logfile, "l", "", "alser log file name")
	flag.BoolVar(&options.lock, "lock", true, "lock so that only 1 instance can run")
	flag.BoolVar(&options.showversion, "version", false, "show version")
	flag.BoolVar(&options.showparsers, "parsers", false, "show all parsers")
	flag.BoolVar(&options.debug, "debug", false, "debug mode")
	flag.BoolVar(&options.daemon, "daemon", false, "run as daemon")
	flag.BoolVar(&options.test, "test", false, "test mode")
	flag.IntVar(&options.tick, "t", TICKER, "tick interval in seconds")
	flag.BoolVar(&options.tailmode, "tail", true, "tail mode")
	flag.BoolVar(&options.dryrun, "dryrun", false, "dry run")
	flag.StringVar(&options.cpuprof, "cpuprof", "", "cpu profiling file")
	flag.StringVar(&options.memprof, "memprof", "", "memory profiling file")
	flag.StringVar(&options.parser, "parser", "", "only run this parser id")
	flag.StringVar(&options.locale, "locale", "", "only guard this locale")

	flag.Usage = showUsage

	flag.Parse()
	if options.debug {
		options.verbose = true
	}
}

func showUsage() {
	fmt.Fprint(os.Stderr, USAGE)
	flag.PrintDefaults()
}

func showVersionAndExit() {
	fmt.Fprintf(os.Stderr, "ALSer %s (build: %s)\n", VERSION, BuildID)
	fmt.Fprintf(os.Stderr, "Built with %s %s for %s/%s\n",
		runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	os.Exit(0)
}

func setupMaxProcsAndProfiler() {
	numCpu := runtime.NumCPU()
	maxProcs := numCpu/2 + 1
	runtime.GOMAXPROCS(numCpu)
	logger.Printf("build[%s] starting with %d/%d CPUs...\n", BuildID, maxProcs, numCpu)

	if options.cpuprof != "" {
		f, err := os.Create(options.cpuprof)
		if err != nil {
			panic(err)
		}

		logger.Printf("CPU profiler %s enabled\n", options.cpuprof)
		pprof.StartCPUProfile(f)
	}

	if options.memprof != "" {
		f, err := os.Create(options.cpuprof)
		if err != nil {
			panic(err)
		}

		logger.Printf("CPU profiler %s enabled\n", options.cpuprof)
		pprof.WriteHeapProfile(f)
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

	logOptions := LOG_OPTIONS
	if options.debug {
		logOptions = LOG_OPTIONS_DEBUG
	}

	return log.New(logWriter, fmt.Sprintf("[%d]", os.Getpid()), logOptions)
}
