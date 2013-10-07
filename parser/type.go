package parser

import (
	"time"
)

type Parser interface {
	parseLine(line string)
	getStats(duration time.Duration)
}
