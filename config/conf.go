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
	Name    string
	Type    string // float, string(default), int
	Contain string
	Ignores []string
	Visible bool
	Regex   []string
}

type ConfParser struct {
	Id      string
	Class   string
	Title   string
	Enabled bool
	Keys    []LineKey // besides area,ts
	Colors  []string  // fg, effects, bg

	PrintFormat string // printf
	ShowSummary bool

	MailRecipients    []string
	MailSubjectPrefix string

	Sleep         int
	BeepThreshold int

	DbName      string //db name is table name
	CreateTable string
	InsertStmt  string
	StatsStmt   string
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
		parser.PrintFormat = this.String(keyPrefix+"printf", "")
		parser.Title = this.String(keyPrefix+"title", "")
		parser.BeepThreshold = this.Int(keyPrefix+"beep_threshold", 0)
		parser.Sleep = this.Int(keyPrefix+"sleep", 10)
		parser.Colors = this.StringList(keyPrefix+"colors", nil)
		parser.DbName = this.String(keyPrefix+"dbname", "")
		parser.CreateTable = this.String(keyPrefix+"create_table", "")
		parser.InsertStmt = this.String(keyPrefix+"insert_stmt", "")
		parser.StatsStmt = this.String(keyPrefix+"stats_stmt", "")
		parser.ShowSummary = this.Bool(keyPrefix+"summary", false)
		parser.Enabled = this.Bool(keyPrefix+"enabled", true)

		// keys
		keys := this.List(keyPrefix+"keys", nil)
		if keys != nil {
			for j := 0; j < len(keys); j++ {
				prefix := fmt.Sprintf("%s[%d].", keyPrefix+"keys", j)
				key := LineKey{}
				key.Name = this.String(prefix+"name", "")
				key.Type = this.String(prefix+"type", "string")
				key.Contain = this.String(prefix+"contain", "")
				key.Ignores = this.StringList(prefix+"ignores", nil)
				key.Regex = this.StringList(prefix+"regex", nil)
				key.Visible = this.Bool(prefix+"visible", true)
				parser.Keys = append(parser.Keys, key)
			}
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

func (this *Config) ParserById(id string) *ConfParser {
	for _, p := range this.Parsers {
		if p.Id == id {
			return &p
		}
	}

	return nil
}

func (this *ConfParser) MailEnabled() bool {
	return len(this.MailRecipients) > 0
}

func (this *ConfParser) MailTos() string {
	return strings.Join(this.MailRecipients, ",")
}

func (this *ConfParser) StatsSql() string {
	return fmt.Sprintf(this.StatsStmt, this.DbName)
}

func (this *ConfGuard) HasParser(parser string) bool {
	for _, p := range this.Parsers {
		if p == parser {
			return true
		}
	}

	return false
}
