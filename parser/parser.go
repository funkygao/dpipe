package parser

import (
    json "github.com/bitly/go-simplejson"
    "time"
)

type Parser interface {
    ParseLine(line string) (area string, ts uint64, data *json.Json)
    GetStats(duration time.Duration)
}
