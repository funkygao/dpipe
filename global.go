package main

import (
	"github.com/funkygao/funpipe/engine"
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
	USAGE = `funpipe

Flags:
`
)
