package parser

import (
	"github.com/funkygao/alser/config"
	"strings"
)

// area,ts,....,hostIp
type HostLineParser struct {
	AlsParser
}

func newHostLineParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *HostLineParser) {
	this = new(HostLineParser)
	this.init(conf, chUpstream, chDownstream)

	return
}

func (this *HostLineParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.AlsParser.ParseLine(line)

	parts := strings.Split(msg, ",")
	n := len(parts)
	host, data := parts[n-1], strings.Join(parts[:n-1], ",")

	this.colorPrintfLn("%3s %15s %s", area, host, data)
	this.beep()

	return
}
