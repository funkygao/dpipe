package parser

import (
	"github.com/funkygao/alser/config"
)

// pid: only run this single parser id
func InitParsers(pid string, conf *config.Config, chUpstreamAlarm chan<- Alarm) {
	for _, g := range conf.Guards {
		for _, parserId := range g.Parsers {
			if pid != "" && pid != parserId {
				continue
			}

			if _, present := allParsers[parserId]; present {
				continue
			}

			confParser := conf.ParserById(parserId)
			allParsers[parserId] = createParser(confParser, chUpstreamAlarm, chParserAlarm)
		}
	}
}

// Dispatch a line of log entry to target parser by name
func Dispatch(parserName, line string) {
	if p, present := allParsers[parserName]; present {
		p.ParseLine(line)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
