package parser

import (
	"time"
)

type Parser interface {
	ParseLine(line string)
	GetStats(duration time.Duration)
}
