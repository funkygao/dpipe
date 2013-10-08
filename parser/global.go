package parser

import (
    "log"
)

var (
    logger     *log.Logger
    allParsers map[string]Parser
    verbose    bool
    debug      bool
)

const (
    LINE_SPLITTER  = ","
    LINE_SPLIT_NUM = 3
)

// Pass through logger
func SetLogger(l *log.Logger) {
    logger = l
}

// Enable/disable debug mode
func SetDebug(d bool) {
    debug = d
}

// Enable verbose or not
func SetVerbose(v bool) {
    verbose = v
}
