/*
Configurations shared between alser and parers.

        Rule
          |
     +---------+
     |         |
  Project   Project
               |
          +---------------+
          |               |
     []DataSource     []Parser
                          |
                      +-------+
                      |       |
                     Key     Key

*/
package config

import (
	"errors"
	"fmt"
	conf "github.com/daviddengcn/go-ljson-conf"
	"regexp"
	"strings"
)

const (
	DATASOURCE_DB   = "db"
	DATASOURCE_FILE = "file"
	DATASOURCE_SYS  = "sys"

	INDEX_YEARMONTH = "@ym"
)

type Config struct {
	*conf.Conf
	Guards  []ConfGuard
	Parsers []ConfParser
}

type ConfGuard struct {
	Enabled        bool   // enabled
	Type           string // type
	TailLogGlob    string // tail_glob
	HistoryLogGlob string // history_glob
	Tables         string // sql like grammer, e,g. log_%

	Parsers []string
}

// Key data sink to 4 kinds of targets
// ======== ========== ==============
// sqldb(d) indexer(i) sink
// ======== ========== ==============
//        Y Y          3, default
//        Y N          2, alarm only
//        N Y          1, index only
//        N N          0, validator only
// ======== ========== ==============
type LineKey struct {
	Name    string
	Type    string // float, string(default), int, money
	Contain string // only being validator instead of data
	Ignores []string
	Filters []string // currently not used yet TODO
	Regex   []string
	Sink    int // bit op
}

type ConfParser struct {
	Id           string
	Class        string
	Title        string
	MsgRegex     string
	MsgRegexKeys []string
	Enabled      bool
	Keys         []LineKey // besides area,ts
	Colors       []string  // fg, effects, bg

	PrintFormat   string // printf
	InstantFormat string // instantf, echo for each occurence
	ShowSummary   bool
	Indexing      bool
	IndexName     string
	IndexAll      bool // index all keys, we needn't define keys in rules
	LevelRange    []int

	Sleep           int
	BeepThreshold   int
	AbnormalPercent float64
	AbnormalBase    int

	DbName      string //db name is table name
	CreateTable string
	InsertStmt  string
	StatsStmt   string
	PersistDb   string // will never auto delete for manual analytics
}

func LoadRuleEngine(fn string) (*Config, error) {
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
		parser.PrintFormat = this.String(keyPrefix+"printf", "")
		parser.InstantFormat = this.String(keyPrefix+"instantf", "")
		parser.Title = this.String(keyPrefix+"title", "")
		parser.MsgRegex = this.String(keyPrefix+"msg_regex", "")
		parser.MsgRegexKeys = this.StringList(keyPrefix+"msg_regex_keys", nil)
		parser.BeepThreshold = this.Int(keyPrefix+"beep_threshold", 0)
		parser.Sleep = this.Int(keyPrefix+"sleep", 10)
		parser.Colors = this.StringList(keyPrefix+"colors", nil)
		parser.DbName = this.String(keyPrefix+"dbname", "")
		parser.PersistDb = this.String(keyPrefix+"persistdb", "")
		parser.CreateTable = this.String(keyPrefix+"create_table", "")
		parser.InsertStmt = this.String(keyPrefix+"insert_stmt", "")
		parser.StatsStmt = this.String(keyPrefix+"stats_stmt", "")
		parser.ShowSummary = this.Bool(keyPrefix+"summary", false)
		parser.Indexing = this.Bool(keyPrefix+"indexing", true)
		parser.IndexAll = this.Bool(keyPrefix+"indexall", false)
		parser.LevelRange = this.IntList(keyPrefix+"lvrange", nil)
		parser.IndexName = this.String(keyPrefix+"indexname", INDEX_YEARMONTH)
		parser.Enabled = this.Bool(keyPrefix+"enabled", true)
		parser.AbnormalPercent = this.Float(keyPrefix+"abnormal_percent", 1.5)
		parser.AbnormalBase = this.Int(keyPrefix+"abnormal_base", 10)

		// keys
		keys := this.List(keyPrefix+"keys", nil)
		if keys != nil {
			for j := 0; j < len(keys); j++ {
				prefix := fmt.Sprintf("%s[%d].", keyPrefix+"keys", j)
				key := LineKey{}
				key.Name = this.String(prefix+"name", "")
				key.Type = this.String(prefix+"type", "string")
				key.Contain = this.String(prefix+"contain", "")
				key.Sink = this.Int(prefix+"sink", 3)
				key.Ignores = this.StringList(prefix+"ignores", nil)
				key.Filters = this.StringList(prefix+"filters", nil)
				key.Regex = this.StringList(prefix+"regex", nil)
				parser.Keys = append(parser.Keys, key)

				if key.Contain != "" {
					// validator only, will never sink to db or indexer
					key.Sink = 0
				}
			}
		}

		this.Parsers = append(this.Parsers, parser)
	}

	// guards section
	guards := this.List("guards", nil)
	for i := 0; i < len(guards); i++ {
		keyPrefix := fmt.Sprintf("guards[%d].", i)
		guard := ConfGuard{}
		guard.Enabled = this.Bool(keyPrefix+"enabled", true)
		guard.Type = this.String(keyPrefix+"type", DATASOURCE_FILE)
		guard.TailLogGlob = this.String(keyPrefix+"tail_glob", "")
		guard.HistoryLogGlob = this.String(keyPrefix+"history_glob", "")
		guard.Parsers = this.StringList(keyPrefix+"parsers", nil)
		guard.Tables = this.String(keyPrefix+"tables", "")
		if guard.Type != DATASOURCE_SYS {
			if guard.Tables != "" && (guard.TailLogGlob != "" || guard.HistoryLogGlob != "") {
				return nil, errors.New("can't have both file and db as datasource")
			}
			if guard.Tables == "" && guard.TailLogGlob == "" && guard.HistoryLogGlob == "" {
				return nil, errors.New("non datasource defined")
			}
		}

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

func (this *Config) CountOfGuards() (c int) {
	for _, g := range this.Guards {
		if g.Enabled {
			c += 1
		}
	}

	return
}

func (this *Config) ParserById(id string) *ConfParser {
	for _, p := range this.Parsers {
		if p.Id == id {
			return &p
		}
	}

	return nil
}

func (this *ConfParser) StatsSql() string {
	return fmt.Sprintf(this.StatsStmt, this.DbName)
}

func (this *ConfParser) LineKeyByName(name string) (lineKey LineKey, err error) {
	for _, lk := range this.Keys {
		if lk.Name == name {
			return lk, nil
		}
	}

	return LineKey{Name: ""}, errors.New("not found")
}

func (this *ConfGuard) HasParser(parser string) bool {
	for _, p := range this.Parsers {
		if p == parser {
			return true
		}
	}

	return false
}

func (this *LineKey) MsgIgnored(msg string) bool {
	for _, ignore := range this.Ignores {
		if strings.Contains(msg, ignore) {
			return true
		}

		if strings.HasPrefix(ignore, "regex:") {
			pattern := strings.TrimSpace(ignore[6:])
			// TODO lessen the overhead
			if matched, err := regexp.MatchString(pattern, msg); err == nil && matched {
				return true
			}
		}
	}

	// filters means only when the key satisfy at least one of the filter rule
	// will the msg be accepted
	if this.Filters != nil {
		for _, f := range this.Filters {
			if msg == f {
				return false
			}
		}

		return true
	}

	return false
}

func (this *LineKey) Alarmable() bool {
	return this.Sink&2 != 0
}

func (this *LineKey) Indexable() bool {
	return this.Sink&1 != 0
}
