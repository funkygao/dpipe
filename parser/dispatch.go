package parser

import (
	"github.com/funkygao/alser/config"
)

// parser: only run this single parser
func InitParsers(parser string, conf *config.Config, chAlarm chan<- parser.Alarm) {
	for _, g := range conf.Guards {
		for _, parserId := range g.Parsers {
			if parser != "" && parser != parserId {
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
