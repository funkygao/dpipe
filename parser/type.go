package parser

import (
	"time"
)

type Parser interface {
	ParseLine(line string)
	GetStats(duration time.Duration)
}

type logData map[string]interface {}
