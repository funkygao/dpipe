package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

type Option struct {
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

// parse argv to Option struct
func parseFlags() *Option {
	var (
		verbose     = flag.Bool("v", false, "verbose")
		config      = flag.String("c", "etc/alser.cf", "config json file")
		logfile     = flag.String("l", "", "alser log file name")
		lock        = flag.Bool("lock", true, "lock so that only 1 instance can run")
		showversion = flag.Bool("version", false, "show version")
		showparsers = flag.Bool("parsers", false, "show all parsers")
		debug       = flag.Bool("debug", false, "debug mode")
		daemon      = flag.Bool("daemon", false, "run as daemon")
		test        = flag.Bool("test", false, "test mode")
		tick        = flag.Int("t", TICKER, "tick interval in seconds")
		tailmode    = flag.Bool("tail", true, "tail mode")
		dryrun      = flag.Bool("dryrun", false, "dry run")
		cpuprof     = flag.String("cpuprof", "", "cpu profiling file")
		parser      = flag.String("parser", "", "only run this parser")
		locale      = flag.String("locale", "", "only guard this locale")
	)
	flag.Usage = showUsage

	flag.Parse()

	return &Option{*verbose, *config, *showversion, *logfile, *debug,
		*test, *tick, *tailmode, *dryrun, *cpuprof, *parser, *locale, *lock, *daemon,
		*showparsers}
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
	logger.Printf("starting with %d/%d CPUs...\n", maxProcs, numCpu)
}
