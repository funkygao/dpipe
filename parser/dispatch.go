package parser

import (
	"github.com/funkygao/alser/config"
)

func createParser(conf *config.ConfParser, chUpstreamAlarm chan<- Alarm, chDownstreamAlarm chan<- string) Parser {
	mutex.Lock()
	defer mutex.Unlock()

	if conf.Class == "JsonLineParser" {
		return newJsonLineParser(conf, chUpstreamAlarm, chDownstreamAlarm)
	} else if conf.Class == "HostLineParser" {
		return newHostLineParser(conf, chUpstreamAlarm, chDownstreamAlarm)
	} else if conf.Class == "RegexCollectorParser" {
		return newRegexCollectorParser(conf, chUpstreamAlarm, chDownstreamAlarm)
	}

	return newJsonCollectorParser(conf, chUpstreamAlarm, chDownstreamAlarm)
}

// pid: only run this single parser id
func InitParsers(pid string, conf *config.Config, chUpstreamAlarm chan<- Alarm) {
	go runSendAlarmsWatchdog(conf)

	indexer = newIndexer(conf)
	go indexer.mainLoop()

	for _, g := range conf.Guards {
		for _, parserId := range g.Parsers {
			if pid != "" && pid != parserId {
				continue
			}

			if _, present := allParsers[parserId]; present {
				continue
			}

			confParser := conf.ParserById(parserId)
			if confParser == nil {
				panic("invalid parser id: " + parserId)
			}

			if debug {
				logger.Printf("create parser[%s] for %s\n", parserId, g.TailLogGlob)
			}

			allParsers[parserId] = createParser(confParser, chUpstreamAlarm, chParserAlarm)
		}
	}
}

// Dispatch a line of log entry to target parser by name
func Dispatch(parserId, line string) {
	if p, present := allParsers[parserId]; present {
		p.ParseLine(line)
	}
}
