package parser

import (
	"log"
)

func init() {
	allParsers = make(map[string] Parser)

	allParsers["DefaultParser"] = DefaultParser{}
	allParsers["MemcacheFailParser"] = MemcacheFailParser{}
}

func SetLogger(l *log.Logger) {
	logger = l
}

func SetDebug(d bool) {
	debug = d
}

func SetVerbose(v bool) {
	verbose = v
}
