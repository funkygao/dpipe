package parser

import (
	"github.com/funkygao/alser/config"
	"strconv"
	"strings"
)

// area,ts,....,hostIp
type HostLineParser struct {
	CollectorParser
}

func newHostLineParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *HostLineParser) {
	this = new(HostLineParser)
	this.init(conf, chUpstream, chDownstream)

	go this.CollectAlarms()

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

	// ignores(cons: key name must be 'data')
	if key, err := this.conf.LineKeyByName("data"); err == nil && key.Ignores != nil {
		for _, ignore := range key.Ignores {
			if strings.Contains(data, ignore) {
				return
			}
		}
	}

	// syslog-ng als handling statastics
	parts = strings.Split(msg, "Log statistics; ")
	if len(parts) == 2 {
		// it is syslog-ng entry
		rawStats := parts[1]

		// dropped parsing
		dropped := syslogngDropped.FindAllStringSubmatch(rawStats, 10000)
		for _, d := range dropped {
			num := d[2]
			if num == "0" {
				continue
			}

			// 丢东西啦，立刻报警
			this.alarmf("%3s %s dropped %s", area, d[1], num)
			this.colorPrintfLn("%3s %s dropped %s", area, d[1], num)
			this.beep()
		}

		// processed parsing
		processed := syslogngProcessed.FindAllStringSubmatch(rawStats, 10000)
		for _, p := range processed {
			val, err := strconv.Atoi(p[2])
			if err != nil || val == 0 {
				continue
			}

			this.insert(p[1], val)
		}

		return
	}

	this.colorPrintfLn("%3s %15s %s", area, host, data)
	if this.conf.BeepThreshold > 0 {
		this.beep()
	}

	return
}
