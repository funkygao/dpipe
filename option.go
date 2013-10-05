package main

import (
	//"encoding/json"
	"flag"
)

type Option struct {
	verbose bool
	config string
	showversion bool
	logfile string
	debug bool
}

// load json conf file into struct
func (this *Option) loadConf() {

}

func ParseFlags() (*Option, error) {
	var (
		verbose = flag.Bool("v", false, "verbose")
		config = flag.String("c", "conf/alser.json", "config json file")
		logfile = flag.String("l", "", "log file name")
		showversion = flag.Bool("version", false, "show version")
		debug = flag.Bool("debug", false, "debug mode")
	)
	flag.Parse()

	option := &Option{*verbose, *config, *showversion, *logfile, *debug}
	return option, nil
}
