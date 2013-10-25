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
package alsparser

import (
	json "github.com/bitly/go-simplejson"
)

// Parser prototype
type Parser interface {
	ParseLine(line string) (area string, ts uint64, data *json.Json)
	Stop()
}

// Create all parsers by name at once
func NewParsers(parsers []string, chAlarm chan<- Alarm) {
	for _, p := range parsers {
		switch p {
		case "MemcacheFailParser":
			allParsers["MemcacheFailParser"] = newMemcacheFailParser("MemcacheFailParser", chAlarm)

		case "ErrorLogParser":
			allParsers["ErrorLogParser"] = newErrorLogParser("ErrorLogParser", chAlarm,
				"var/error.sqlite", ERRLOG_CREATE_TABLE, ERRLOG_INSERT)

		case "PaymentParser":
			allParsers["PaymentParser"] = newPaymentParser("PaymentParser", chAlarm,
				"var/payment.sqlite", PAYMENT_CREATE_TABLE, PAYMENT_INSERT)

		case "PhpErrorLogParser":
			allParsers["PhpErrorLogParser"] = newPhpErrorLogParser("PhpErrorLogParser", chAlarm,
				"var/phperror.sqlite", PHPERROR_CREATE_TABLE, PHPERROR_INSERT)

		case "SlowResponseParser":
			allParsers["SlowResponseParser"] = newSlowResponseParser("SlowResponseParser", chAlarm,
				"var/slowresp.sqlite", SLOWRESP_CREATE_TABLE, SLOWRESP_INSERT)

		case "LevelUpParser":
			allParsers["LevelUpParser"] = newLevelUpParser("LevelUpParser", chAlarm,
				"var/levelup.sqlite", LEVELUP_CREATE_TABLE, LEVELUP_INSERT)

		default:
			logger.Println("Invalid parser:", p)
		}
	}
}

// Stop all parsers and they will do their cleanup automatically
func StopAll() {
	for _, parser := range allParsers {
		parser.Stop()
	}
}
