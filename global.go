package main

import (
	"log"
	"time"
)

var (
	options *Option
	logger  *log.Logger
	ticker  *time.Ticker

	BuildID = "unknown" // git version id, passed in from shell
)

const (
	VERSION = "0.3.stable"
	AUTHOR  = "gaopeng"

	LOG_OPTIONS       = log.Ldate | log.Ltime
	LOG_OPTIONS_DEBUG = log.Ldate | log.Lshortfile | log.Ltime | log.Lmicroseconds

	USAGE = `alser - FunPlus ALS log guard

Flags:
`

	LOCKFILE  = "var/alser.lock"
	TICKER    = 60 * 2 // 2 minutes
	tailSleep = 1      // 1 seconds between tail reading
)
