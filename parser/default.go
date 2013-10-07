package parser

import (
	"encoding/json"
	"strings"
	"time"
	"strconv"
)

func (this DefaultParser) ParseLine(line string) (area string, ts uint64, data logData) {
	fields := strings.SplitN(line, LINE_SPLITTER, LINE_SPLIT_NUM)

	area = fields[0]
	var err error
	ts, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal([]byte(fields[2]), &data); err != nil {
		panic(err)
	}

	if debug && verbose {
		logger.Printf("%s %#v", area, data)
	}

	//bb := make(map[string]interface {})
	//json.Unmarshal([]byte(data["_log_info"]), &bb)
	return
}

func (this DefaultParser) GetStats(duration time.Duration) {

}
