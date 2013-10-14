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
	pprof       string
	parser      string
}

func (this *Option) showVersionOnly() bool {
	return this.showversion
}

func (this *Option) validate() {
	if this.showVersionOnly() {
		fmt.Fprintf(os.Stderr, "ALSer %s (build: %s)\n", VERSION, BuildID)
		fmt.Fprintf(os.Stderr, "Built with %s %s for %s/%s\n",
			runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

// parse argv to Option struct
func parseFlags() *Option {
	var (
		verbose     = flag.Bool("v", false, "verbose")
		config      = flag.String("c", "conf/alser.json", "config json file")
		logfile     = flag.String("l", "", "alser log file name")
		showversion = flag.Bool("version", false, "show version")
		debug       = flag.Bool("debug", false, "debug mode")
		test        = flag.Bool("test", false, "test mode")
		t           = flag.Int("t", tick, "tick interval in seconds")
		tailmode    = flag.Bool("tail", false, "tail mode")
		dr          = flag.Bool("dry-run", false, "dry run")
		cpuprof     = flag.String("pprof", "", "cpu pprof file")
		p           = flag.String("parser", "", "only run this parser class")
	)
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	return &Option{*verbose, *config, *showversion, *logfile, *debug,
		*test, *t, *tailmode, *dr, *cpuprof, *p}
}
