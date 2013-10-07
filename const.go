package main

import (
	"log"
)

const (
	version = "0.1.a"
	author = "gaopeng"
)

const (
	LOG_OPTIONS = log.LstdFlags | log.Lshortfile
	LOG_OPTIONS_DEBUG = log.Ldate | log.Lshortfile | log.Ltime | log.Lmicroseconds
)

const (
	usage = `alser - FunPlus ALS log guard

Flags:
`
)
