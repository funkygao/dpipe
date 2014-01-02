package main

import (
	"github.com/funkygao/funpipe/engine"
	"log"
)

var (
	globals *engine.GlobalConfigStruct

	BuildID = "unknown" // git version id, passed in from shell

	options struct {
		verbose     bool
		configfile  string
		showversion bool
		logfile     string
		debug       bool
		tick        int
		dryrun      bool
		cpuprof     string
		memprof     string
		lockfile    string
	}
)

const (
	LOG_OPTIONS       = log.Ldate | log.Ltime
	LOG_OPTIONS_DEBUG = log.Ldate | log.Lshortfile | log.Ltime | log.Lmicroseconds

	USAGE = `funpipe

Flags:
`
)
