/*
Configurations shared between alser and parers.
*/
package config

import (
	"errors"
	"fmt"
	conf "github.com/daviddengcn/go-ljson-conf"
	"strings"
)

type ConfGuard struct {
	TailLogGlob    string
	HistoryLogGlob string
	Parsers        []string
}

type LineKey struct {
	Key      string
	Required bool
	Show     bool
}

type ConfParser struct {
	Id                string
	Class             string
	Keys              []LineKey // by line
	Colors            []string  // fg, effects, bg
	MailRecipients    []string
	MailSubjectPrefix string
}

type Config struct {
	*conf.Conf
	Guards  []ConfGuard
	Parsers []ConfParser
}

func LoadConfig(fn string) (*Config, error) {
	cf, err := conf.Load(fn)
	if err != nil {
		return nil, err
	}

	this := new(Config)
	this.Conf = cf
	this.Guards = make([]ConfGuard, 0)
	this.Parsers = make([]ConfParser, 0)

	// parsers section
	parsers := this.List("parsers", nil)
	for i := 0; i < len(parsers); i++ {
		keyPrefix := fmt.Sprintf("parsers[%d].", i)
		parser := ConfParser{}
		parser.Id = this.String(keyPrefix+"id", "")
		parser.Class = this.String(keyPrefix+"class", "")
		parser.MailRecipients = this.StringList(keyPrefix+"mail_recipents", nil)
		parser.MailSubjectPrefix = this.String(keyPrefix+"mail_subject_prefix", "")
		parser.Colors = this.StringList(keyPrefix+"colors", nil)
		// keys
		keys := this.List(keyPrefix+"keys", nil)
		if keys == nil {
			return nil, errors.New("keys can't be empty in parser")
		}
		for j := 0; j < len(keys); j++ {
			prefix := fmt.Sprintf("%s[%d].", keyPrefix+"keys", j)
			key := LineKey{}
			key.Key = this.String(prefix+"key", "")
			key.Required = this.Bool(prefix+"required", false)
			key.Show = this.Bool(prefix+"show", false)
			parser.Keys = append(parser.Keys, key)
		}

		this.Parsers = append(this.Parsers, parser)
	}

	// guards section
	guards := this.List("guards", nil)
	for i := 0; i < len(guards); i++ {
		keyPrefix := fmt.Sprintf("guards[%d].", i)
		guard := ConfGuard{}
		guard.TailLogGlob = this.String(keyPrefix+"tail_glob", "")
		guard.HistoryLogGlob = this.String(keyPrefix+"history_glob", "")
		guard.Parsers = this.StringList(keyPrefix+"parsers", nil)

		this.Guards = append(this.Guards, guard)
	}

	// validation
	if this.hasDupParsers() {
		return nil, errors.New("has dup parsers")
	}

	return this, nil
}

func (this *Config) IsParserApplied(parser string) bool {
	for _, g := range this.Guards {
		for _, p := range g.Parsers {
			if p == parser {
				return true
			}
		}
	}

	return false
}

// Dup parser id
func (this *Config) hasDupParsers() bool {
	parsers := make(map[string]bool)
	for _, p := range this.Parsers {
		if _, present := parsers[p.Id]; present {
			return true
		}

		parsers[p.Id] = true
	}

	return false
}

func (this *ConfParser) MailEnabled() bool {
	return len(this.MailRecipients) > 0
}

func (this *ConfParser) MailTos() string {
	return strings.Join(this.MailRecipients, ",")
}
