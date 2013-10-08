package parser

import (
    json "github.com/bitly/go-simplejson"
    "strconv"
    "strings"
    "time"
)


// Parent parser for all
type DefaultParser struct {
}

func (this DefaultParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
    fields := strings.SplitN(line, LINE_SPLITTER, LINE_SPLIT_NUM)

    area = fields[0]
    var err error
    ts, err = strconv.ParseUint(fields[1], 10, 64)
    if err != nil {
        panic(err)
    }
    ts /= 1000

    data, err = json.NewJson([]byte(fields[2]))
    if err != nil {
        panic(err)
    }

    return
}

func (this DefaultParser) GetStats(duration time.Duration) {

}
