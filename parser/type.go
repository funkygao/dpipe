package parser

import (
	"time"
	json "github.com/bitly/go-simplejson"
)

type Parser interface {
	ParseLine(line string) (area string, ts uint64, data *json.Json)
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
