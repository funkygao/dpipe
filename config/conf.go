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
	Name     string
	Type     string // float, string, int
	Required bool
	Show     bool
	Groupby  bool
	MustBe   string
}

type ConfParser struct {
	Id          string
	Class       string
	Keys        []LineKey // by line
	Colors      []string  // fg, effects, bg
	PrintFormat string    // printf
	ShowSrcHost bool      // show log src host
	ShowUri     bool
	Title       string

	MailRecipients    []string
	MailSubjectPrefix string

	Sleep         int
	BeepThreshold int

	DbName      string //db name is table name
	CreateTable string
	InsertStmt  string
	SumColumn   string
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
		parser.ShowUri = this.String(keyPrefix+"show_uri", false)
		parser.ShowSrcHost = this.String(keyPrefix+"show_host", true)
		parser.BeepThreshold = this.Int(keyPrefix+"beep_threshold", 0)
		parser.Sleep = this.Int(keyPrefix+"sleep", 10)
		parser.Colors = this.StringList(keyPrefix+"colors", nil)
		parser.DbName = this.String(keyPrefix+"dbname", "")
		parser.CreateTable = this.String(keyPrefix+"create_table", "")
		parser.InsertStmt = this.String(keyPrefix+"insert_stmt", "")
		parser.SumColumn = this.String(keyPrefix+"sum", "")

		// keys
		keys := this.List(keyPrefix+"keys", nil)
		if keys == nil {
			return nil, errors.New("keys can't be empty in parser")
		}
		for j := 0; j < len(keys); j++ {
			prefix := fmt.Sprintf("%s[%d].", keyPrefix+"keys", j)
			key := LineKey{}
			key.Name = this.String(prefix+"name", "")
			key.Required = this.Bool(prefix+"required", false)
			key.Show = this.Bool(prefix+"show", false)
			key.Groupby = this.Bool(prefix+"groupby", false)
			key.Type = this.Bool(prefix+"type", "string")
			key.MustBe = this.Bool(prefix+"must_be", "")
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

func (this *ConfParser) LogInfoNeeded() bool {
	return this.ShowSrcHost || this.ShowUri
}

func (this *ConfParser) groupByKeys() []string {
	keys := make([]string, 0)

	for _, k := range this.Keys {
		if k.Groupby {
			keys = append(keys, k.Name)
		}
	}

	return keys
}

func (this *ConfParser) StatsSql() string {
	groupBys := strings.Join(this.gr, ",")
	if this.SumColumn != "" {
		return fmt.Printf("SELECT SUM(%s) AS c, %s FROM %s WHERE ts<=? GROUP BY %s ORDER BY c DESC",
			this.SumColumn,
			groupBys, this.DbName, groupBys)
	}
	return fmt.Printf("SELECT COUNT(*) AS c, %s FROM %s WHERE ts<=? GROUP BY %s ORDER BY c DESC",
		groupBys, this.DbName, groupBys)
}

func (this *ConfGuard) HasParser(parser string) bool {
	for _, p := range this.Parsers {
		if p == parser {
			return true
		}
	}

	return false
}
