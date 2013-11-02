package parser

import (
	"github.com/funkygao/alser/config"
)

// pid: only run this single parser id
func InitParsers(pid string, conf *config.Config, chAlarm chan<- Alarm) {
	for _, g := range conf.Guards {
		for _, parserId := range g.Parsers {
			if pid != "" && pid != parserId {
				continue
			}

			confParser := conf.ParserById(parserId)
			allParsers[parserId] = createParser(confParser, chAlarm)
		}
	}
}

// Dispatch a line of log entry to target parser by name
func Dispatch(parserName, line string) {
	p, ok := getParser(parserName)
	if !ok {
		return
	}

	p.ParseLine(line)
}

// Get a parser instance by name
func getParser(parserName string) (p Parser, ok bool) {
	p, ok = allParsers[parserName]
	return
}
