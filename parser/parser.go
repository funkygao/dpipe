/*
Parser for some special messages
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

func Parse(typ string, msg string) (string, error) {
	if typ == "" {
		return "", ErrInvaidParser
	}

	parse, present := allParsers[typ]
	if !present {
		return "", ErrInvaidParser
	}

	return parse(msg), nil
}
