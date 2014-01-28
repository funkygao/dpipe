/*
Parser for some special message fields
Currently soly for alarm
*/
package parser

import (
	"errors"
)

type ParseMsg func(msg string) string

var (
	ErrInvaidParser = errors.New("invalid parser type")

	allParsers = map[string]ParseMsg{
		"syslogngStats": parseSyslogNgStats,
	}
)

func Parse(typ string, msg string) (alarm string, err error) {
	if typ == "" {
		return "", ErrInvaidParser
	}

	parse, present := allParsers[typ]
	if !present {
		return "", ErrInvaidParser
	}

	return parse(msg), nil
}
