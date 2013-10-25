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
	lock        bool
}

// parse argv to Option struct
func parseFlags() *Option {
	var (
		verbose     = flag.Bool("v", false, "verbose")
		config      = flag.String("c", "conf/alser.json", "config json file")
		logfile     = flag.String("l", "", "alser log file name")
		lock        = flag.Bool("lock", true, "lock so that only 1 instance can run")
		showversion = flag.Bool("version", false, "show version")
		debug       = flag.Bool("debug", false, "debug mode")
		test        = flag.Bool("test", false, "test mode")
		t           = flag.Int("t", tick, "tick interval in seconds")
		tailmode    = flag.Bool("tail", false, "tail mode")
		dr          = flag.Bool("dry-run", false, "dry run")
		cpuprof     = flag.String("pprof", "", "cpu pprof file")
		p           = flag.String("p", "", "only run this parser class")
	)
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	return &Option{*verbose, *config, *showversion, *logfile, *debug,
		*test, *t, *tailmode, *dr, *cpuprof, *p, *lock}
}
