package main

import (
	"log"
	"time"
)

var (
	logger *log.Logger

	BuildID = "unknown" // git version id, passed in from shell

	allWorkers map[string]bool // key is datasource name
	ticker     *time.Ticker

	options struct {
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
)

const (
	LOG_OPTIONS       = log.Ldate | log.Ltime
	LOG_OPTIONS_DEBUG = log.Ldate | log.Lshortfile | log.Ltime | log.Lmicroseconds

	USAGE = `alser - FunPlus ALS(application logging system) Guard

Flags:
`

	LOCKFILE = "var/alser.lock"
	TICKER   = 60 * 10 // default ticker, 10 minutes

	FLASHLOG_DSN = "flashlog:flashlog@unix(/var/run/mysqld/mysqld.sock)/flashlog?charset=utf8"
)
