package parser

import (
	"github.com/funkygao/alser/config"
	"log"
)

type Stopable interface {
	Stop()
}

type Waitable interface {
	Wait()
}

// Parser prototype
type Parser interface {
	ParseLine(line string) (area string, ts uint64, msg string)
	Stopable
	Waitable
}

type AlarmCollector interface {
	CollectAlarms()
}

// Pass through logger
func SetLogger(l *log.Logger) {
	logger = l
}

// Enable/disable debug mode
func SetDebug(d bool) {
	debug = d
}

// Enable verbose or not
func SetVerbose(v bool) {
	verbose = v
}

func SetDryRun(dr bool) {
	dryRun = dr
}

func SetDaemon(d bool) {
	daemonize = d
}

func createParser(conf *config.ConfParser, chUpstreamAlarm chan<- Alarm, chDownstreamAlarm chan<- string) Parser {
	mutex.Lock()
	defer mutex.Unlock()

	if conf.Class == "JsonLine" {
		return newJsonLineParser(conf, chUpstreamAlarm, chDownstreamAlarm)
	} else if conf.Class == "DbParser" {
		return newDbParser(conf, chUpstreamAlarm, chDownstreamAlarm)
	}

	return newDbParser(conf, chUpstreamAlarm, chDownstreamAlarm)
}

// Stop all parsers and they will do their cleanup automatically
func StopAll() {
	for _, parser := range allParsers {
		parser.Stop()
	}
}

func WaitAll() {
	for _, parser := range allParsers {
		parser.Wait()
	}

	close(chParserAlarm)
}

func ParsersCount() int {
	return len(allParsers)
}

func Parsers() map[string]Parser {
	return allParsers
}
