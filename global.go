package main

import (
	"log"
	"os"
	"syscall"
	"time"
)

var (
	options *Option
	logger  *log.Logger

	caredSignals = []os.Signal{
		syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT,
		syscall.SIGHUP, syscall.SIGSTOP, syscall.SIGQUIT,
	}

	ticker *time.Ticker

	BuildID = "unknown" // git version id, passed in from shell
)

const (
	VERSION = "0.1.rc"
	AUTHOR  = "gaopeng"
)

const (
	ALARM_OPTIONS     = log.LstdFlags
	LOG_OPTIONS       = log.LstdFlags | log.Lshortfile
	LOG_OPTIONS_DEBUG = log.Ldate | log.Lshortfile | log.Ltime | log.Lmicroseconds
)

const (
	usage = `alser - FunPlus ALS log guard

Flags:
`
	lockfile  = "var/alser.lock"
	alarmlog  = "var/alarm.log"
	tick      = 60 * 2 // 2 minutes
	tailSleep = 1      // 1 seconds between tail reading
)
