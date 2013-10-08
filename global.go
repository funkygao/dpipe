package main

import (
    "log"
    "os"
    "syscall"
)

var (
    options *Option
    logger  *log.Logger

    caredSignals = []os.Signal{
        syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT,
        syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT,
    }
)

const (
    version = "0.1.b"
    author  = "gaopeng"
)

const (
    LOG_OPTIONS       = log.LstdFlags | log.Lshortfile
    LOG_OPTIONS_DEBUG = log.Ldate | log.Lshortfile | log.Ltime | log.Lmicroseconds
)

const (
    usage = `alser - FunPlus ALS log guard

Flags:
`
    lockfile = "var/alser.lock"
)
