/*
TODO refactor this ugly monster
too much hidden rules
*/
package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/rule"
	"strconv"
	"strings"
	"time"
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
		if verbose {
			logger.Printf("got empty msg: %s\n", line)
		}

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
		if key.MsgIgnored(data) {
			return
		}
	}

	if dryRun {
		return
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

			if this.conf.Indexing {
				indexJson, _ := json.NewJson([]byte("{}"))
				indexJson.Set(INDEX_COL_AREA, area)
				indexJson.Set(INDEX_COL_TIMESTAMP, ts)
				indexJson.Set("host", host)
				indexJson.Set("ngtyp", p[1])
				indexJson.Set("ngbytes", val)

				date := time.Unix(int64(ts), 0)
				indexer.c <- indexEntry{indexName: this.conf.IndexName, typ: this.conf.Title, date: &date, data: indexJson}
			}

			this.insert(area, ts, p[1], val)
		}

		return
	}

	if this.conf.Indexing {
		indexJson, _ := json.NewJson([]byte("{}"))
		indexJson.Set(INDEX_COL_AREA, area)
		indexJson.Set(INDEX_COL_TIMESTAMP, ts)
		indexJson.Set("host", host)
		indexJson.Set("msg", data)

		date := time.Unix(int64(ts), 0)
		indexer.c <- indexEntry{indexName: this.conf.IndexName, typ: this.conf.Title, date: &date, data: indexJson}
	}

	this.colorPrintfLn("%3s %15s %s", area, host, data)
	this.alarmf("%3s %15s %s", area, host, data)
	if this.conf.BeepThreshold > 0 {
		this.beep()
	}

	return
}
