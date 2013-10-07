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
