package parser

import (
	"time"
)

type logData map[string]interface {}

type Parser interface {
	ParseLine(line string) (area string, ts uint64, data logData)
	GetStats(duration time.Duration)
}

type DefaultParser struct {
	name string
}

type MemcacheFailParser struct {
	DefaultParser
}

type PaymentParser struct {
	DefaultParser
}

type ErrorParser struct {
	DefaultParser
}
