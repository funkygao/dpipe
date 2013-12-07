package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var options struct {
	verbose     bool
	config      string
	showversion bool
	logfile     string
	debug       bool
	test        bool
	tick        int
	tailmode    bool
	dryrun      bool
	cpuprof     string
	parser      string
	locale      string
	lock        bool
	daemon      bool
	showparsers bool
}

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
	flag.StringVar(&options.parser, "parser", "", "only run this parser")
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

func showVersion() {
	fmt.Fprintf(os.Stderr, "ALSer %s (build: %s)\n", VERSION, BuildID)
	fmt.Fprintf(os.Stderr, "Built with %s %s for %s/%s\n",
		runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	os.Exit(0)
}

func setupMaxProcs() {
	numCpu := runtime.NumCPU()
	maxProcs := numCpu/2 + 1
	runtime.GOMAXPROCS(numCpu)
	logger.Printf("build[%s] starting with %d/%d CPUs...\n", BuildID, maxProcs, numCpu)
}
