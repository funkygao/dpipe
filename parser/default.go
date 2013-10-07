package parser

import (
	"encoding/json"
	"strings"
	"time"
)

type DefaultParser struct {
	name string
}

func (this *DefaultParser) Name() string {
	return this.name
}

func (this DefaultParser) ParseLine(line string) {
	fields := strings.SplitN(line, LINE_SPLITTER, 3)
	area, ts, entry := fields[0], fields[1], fields[2]
	var data logData
	json.Unmarshal([]byte(entry), &data)
	logger.Printf("%s %s %#v\n", area, ts, data)
}

func (this DefaultParser) GetStats(duration time.Duration) {

}
