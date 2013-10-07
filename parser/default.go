package parser

import (
	"strings"
	"time"
	json "github.com/bitly/go-simplejson"
	"strconv"
)

func (this DefaultParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	fields := strings.SplitN(line, LINE_SPLITTER, LINE_SPLIT_NUM)

	area = fields[0]
	var err error
	ts, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}

	data, err = json.NewJson([]byte(fields[2]))
	if err != nil {
		panic(err)
	}

	return
}

func (this DefaultParser) GetStats(duration time.Duration) {

}
