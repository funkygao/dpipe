package parser

import (
	"log"
)

var (
	logger *log.Logger
	allParsers map[string] Parser
	verbose bool
)

const (
	LINE_SPLITTER = ","
)
