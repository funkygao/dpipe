package parser

import (
	"time"
	json "github.com/bitly/go-simplejson"
)

type Parser interface {
	ParseLine(line string) (area string, ts uint64, data *json.Json)
	GetStats(duration time.Duration)
}
