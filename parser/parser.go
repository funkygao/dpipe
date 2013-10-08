package parser

import (
    json "github.com/bitly/go-simplejson"
    "time"
)

// Parser prototype
type Parser interface {
    ParseLine(line string, ch chan Alarm) (area string, ts uint64, data *json.Json)
    GetStats(duration time.Duration)
}
