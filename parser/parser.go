/*
               DefaultParser
                   |
       ------------------------------
      |                              |
     DbParser            --------------------
      |                 |
      |          MemcacheFailParser
      |                 
     ---------------------------
    |              |   
   PaymentParser ErrorLogParser

*/
package parser

import (
	json "github.com/bitly/go-simplejson"
	"time"
)

// Parser prototype
type Parser interface {
	ParseLine(line string) (area string, ts uint64, data *json.Json)
	GetStats(duration time.Duration)
	Stop()
}

func NewParsers(parsers []string, chAlarm chan<- Alarm) {
	for _, p := range parsers {
		switch p {
		case "MemcacheFailParser":
			allParsers["MemcacheFailParser"] = newMemcacheFailParser(chAlarm)
		case "ErrorLogParser":
			allParsers["ErrorLogParser"] = newErrorLogParser(chAlarm)
		case "PaymentParser":
			allParsers["PaymentParser"] = newPaymentParser(chAlarm)
		case "PhpErrorLogParser":
			allParsers["PhpErrorLogParser"] = newPhpErrorLogParser(chAlarm)
		default:
			logger.Println("invalid parser:", p)
		}
	}
}

func StopAll() {
	for _, parser := range allParsers {
		parser.Stop()
	}
}
