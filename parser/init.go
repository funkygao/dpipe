package parser

import (
	"log"
)

func init() {
	allParsers = make(map[string] Parser)

	allParsers["DefaultParser"] = DefaultParser{name: "DefaultParser"}
}

func SetLogger(l *log.Logger) {
	logger = l
}

func SetVerbose(v bool) {
	verbose = v
}
