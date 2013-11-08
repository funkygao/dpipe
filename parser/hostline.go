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
	if msg == "" {
		return
	}

	parts := strings.Split(msg, ",")
	n := len(parts)
	host, data := parts[n-1], strings.Join(parts[:n-1], ",")
	if strings.TrimSpace(data) == "" {
		return
	}

	this.colorPrintfLn("%3s %15s %s", area, host, data)
	if this.conf.BeepThreshold > 0 {
		this.beep()
	}

	return
}
