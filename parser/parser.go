/*
Parser for some special message fields
Currently soly for alarm
*/
package parser

import (
	"errors"
)

type ParseMsg func(msg string) (alarm string, severity int)

var (
	ErrInvaidParser = errors.New("invalid parser type")

	allParsers = map[string]ParseMsg{
		"syslogngStats": parseSyslogNgStats,
	}
)

func Parse(typ string, msg string) (alarm string, severity int, err error) {
	if typ == "" {
		return "", 0, ErrInvaidParser
	}

	parse, present := allParsers[typ]
	if !present {
		return "", 0, ErrInvaidParser
	}

	alarm, severity = parse(msg)
	return
}
