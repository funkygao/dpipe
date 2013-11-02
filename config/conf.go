/*
Configurations shared between alser and parers.
*/
package config

import (
	"fmt"
	conf "github.com/daviddengcn/go-ljson-conf"
)

type ConfGuard struct {
	tailLogGlob    string
	historyLogGlob string
	parsers        []string
}

type ConfParser struct {
	id          string
	class       string
	colors      []string // fg, effects, bg
	lineColumns []string
}

type Config struct {
	*conf.Conf
	guards  []ConfGuard
	parsers []ConfParser
}

func LoadConfig(fn string) (*Config, error) {
	cf, err := conf.Load(fn)
	if err != nil {
		return nil, err
	}

	this := new(Config)
	this.Conf = cf
	this.guards = make([]ConfGuard, 0)
	this.parsers = make([]ConfParser, 0)

	// parsers section
	parsers := this.List("parsers", nil)
	for i := 0; i < len(parsers); i++ {
		keyPrefix := fmt.Sprintf("parsers[%d].", i)
		parser := ConfParser{}
		parser.id = this.String(keyPrefix+"id", "")
		parser.class = this.String(keyPrefix+"class", "")
		parser.colors = this.StringList(keyPrefix+"colors", nil)
		parser.lineColumns = this.StringList(keyPrefix+"keys", nil)

		this.parsers = append(this.parsers, parser)
	}

	// guards section
	guards := this.List("guards", nil)
	for i := 0; i < len(guards); i++ {
		keyPrefix := fmt.Sprintf("guards[%d].", i)
		guard := ConfGuard{}
		guard.tailLogGlob = this.String(keyPrefix+"tail_glob", "")
		guard.historyLogGlob = this.String(keyPrefix+"history_glob", "")
		guard.parsers = this.StringList(keyPrefix+"parsers", nil)

		this.guards = append(this.guards, guard)
	}

	return this, nil
}
