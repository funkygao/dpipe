package main

import (
	"log"
)

var (
	options *Option
	logger  *log.Logger

	BuildID = "unknown" // git version id, passed in from shell

	allWorkers map[string]bool // key is logfile name
)

const (
	VERSION = "0.3.stable"
	AUTHOR  = "gaopeng"

	LOG_OPTIONS       = log.Ldate | log.Ltime
	LOG_OPTIONS_DEBUG = log.Ldate | log.Lshortfile | log.Ltime | log.Lmicroseconds

	USAGE = `alser - FunPlus ALS(application logging system) Guard

Flags:
`

	LOCKFILE = "var/alser.lock"
	TICKER   = 60 * 10 // default ticker, 10 minutes
)
