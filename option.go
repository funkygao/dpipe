package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

type Option struct {
	verbose bool
	config string
	showversion bool
	logfile string
	debug bool
	test bool
}

func (this *Option) showVersionOnly() bool {
	return this.showversion
}

func (this *Option) validate() {
	if this.showVersionOnly() {
		cleanup()
		fmt.Fprintf(os.Stderr, "%s %s %s %s\n", "alser", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

// parse argv to Option struct
func parseFlags() (*Option) {
	var (
		verbose = flag.Bool("v", false, "verbose")
		config = flag.String("c", "conf/alser.json", "config json file")
		logfile = flag.String("l", "", "alser log file name")
		showversion = flag.Bool("version", false, "show version")
		debug = flag.Bool("debug", false, "debug mode")
		test = flag.Bool("test", false, "test mode")
	)
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()

		cleanup()
	}

	flag.Parse()

	return &Option{*verbose, *config, *showversion, *logfile, *debug, *test}
}
