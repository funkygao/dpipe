package parser

import (
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
	parts := strings.SplitN(line, ",", 3)
	area, ts, js := parts[0], parts[1], parts[2]
	println(area, ts, js, "go")
}

func (this DefaultParser) GetStats(duration time.Duration) {

}
