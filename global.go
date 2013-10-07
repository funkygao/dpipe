package main

import (
    "log"
)

var (
    options *Option
    logger  *log.Logger
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
