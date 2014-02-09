/*
Parser for some special message fields
Currently soly for alarm
*/
package parser

import (
	"errors"
)

type ParseMsg func(msg string) (match bool, alarm string, severity int)

var (
	ErrInvaidParser = errors.New("invalid parser type")

	allParsers = map[string]ParseMsg{
		"syslogngStats": parseSyslogNgStats,
	}
)

func Parse(typ string, msg string) (match bool, alarm string, severity int, err error) {
	if typ == "" {
		return false, "", 0, ErrInvaidParser
	}

	parse, present := allParsers[typ]
	if !present {
		return false, "", 0, ErrInvaidParser
	}

	match, alarm, severity = parse(msg)
	return
}
