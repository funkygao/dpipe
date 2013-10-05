package main

import (
	"flag"
	"os"
)

type Option struct {
	verbose bool
	config string
	showversion bool
	logfile string
}

func ParseFlags() (*Option, error) {
	var (
		verbose = flag.Bool("v", false, "verbose")
		config = flag.String("c", "conf/alser.json", "config json file")
		logfile = flag.String("l", "", "log file name")
		showversion = flag.Bool("version", false, "show version")
	)
	flag.Parse()

	option := new(Option)
	option.verbose = *verbose
	option.logfile = *logfile
	option.config = *config
	option.showversion = *showversion
	
	return option, nil
}
