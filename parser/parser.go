/*
               AlsParser
                   |
       ------------------------------
      |                              |
     DbParser            --------------------
      |                 |
      |          MemcacheFailParser
      |
     -------------------------------
    |              |                |
   PaymentParser ErrorLogParser PhpErrorLogParser

*/
package parser

import (
	json "github.com/bitly/go-simplejson"
	conf "github.com/daviddengcn/go-ljson-conf"
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
	ParseLine(line string) (area string, ts uint64, data *json.Json)
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

func init() {
	// logger not passed in yet
	if conf, err := conf.Load(CONF_EMAIL); err == nil {
		emailSender = conf.String("sender", "")
		emailHost = conf.String("smtp_host", "")
		emailPasswd = conf.String("passwd", "")

		parserAlarmEnabled = conf.Bool("enabled", true)
		if parserAlarmEnabled {
			go runSendAlarmsWatchdog()
		}
	}

	_, err := conf.Load(CONF_PARSERS)
	checkError(err)

}

// Create all parsers by name at once
func NewParsers(parsers []string, chAlarm chan<- Alarm) {
	for _, p := range parsers {
		NewParser(p, chAlarm)
	}
}

// Create all parsers by name at once
func NewParser(parser string, chAlarm chan<- Alarm) {
	parsersLock.Lock()
	defer parsersLock.Unlock()

	if _, present := allParsers[parser]; present {
		return
	}

	switch parser {
	case "MemcacheFailParser":
		allParsers["MemcacheFailParser"] = newMemcacheFailParser("MemcacheFailParser", FgYellow, chAlarm)

	case "ErrorLogParser":
		allParsers["ErrorLogParser"] = newErrorLogParser("ErrorLogParser", FgRed, chAlarm,
			"var/error.sqlite", "error", ERRLOG_CREATE_TABLE, ERRLOG_INSERT)

	case "MongodbLogParser":
		allParsers["MongodbLogParser"] = newMongodbLogParser("MongodbLogParser", FgCyan+Bright+BgRed, chAlarm,
			"var/mongo.sqlite", "mongo", MONGO_CREATE_TABLE, MONGO_INSERT)

	case "PaymentParser":
		allParsers["PaymentParser"] = newPaymentParser("PaymentParser", FgGreen, chAlarm,
			"var/payment.sqlite", "payment", PAYMENT_CREATE_TABLE, PAYMENT_INSERT)

	case "PhpErrorLogParser":
		allParsers["PhpErrorLogParser"] = newPhpErrorLogParser("PhpErrorLogParser", FgYellow, chAlarm,
			"var/phperror.sqlite", "phperror", PHPERROR_CREATE_TABLE, PHPERROR_INSERT)

	case "SlowResponseParser":
		allParsers["SlowResponseParser"] = newSlowResponseParser("SlowResponseParser", FgBlue, chAlarm,
			"var/slowresp.sqlite", "slowresp", SLOWRESP_CREATE_TABLE, SLOWRESP_INSERT)

	case "LevelUpParser":
		allParsers["LevelUpParser"] = newLevelUpParser("LevelUpParser", FgMagenta, chAlarm,
			"var/levelup.sqlite", "levelup", LEVELUP_CREATE_TABLE, LEVELUP_INSERT)

	default:
		logger.Println("Invalid parser:", parser)
	}

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
