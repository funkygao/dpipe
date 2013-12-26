/*
Configurations shared between alser and parers.
TODO

        Rule
          |
     +---------+---------------+
     |         |               |
  Project   Project     (rule shared across projects)
               |
          +---------------+
          |               |
     []Worker         []Parser
          |               |
     []ParserId       +-------+
                      |       |
                    Field   Field

*/
package rule

import (
	"errors"
	"fmt"
	conf "github.com/daviddengcn/go-ljson-conf"
	"net/url"
)

type RuleEngine struct {
	*conf.Conf

	Workers []ConfWorker
	Parsers []ConfParser
}

// Every worker has 2 modes:
// tail mode and history mode
type ConfWorker struct {
	Enabled     bool   // enabled
	Dsn         string // data source name base
	TailGlob    string // tail_glob
	HistoryGlob string // history_glob

	Parsers []string
}

type ConfParser struct {
	Id           string
	Class        string
	Title        string
	MsgRegex     string
	MsgRegexKeys []string
	Fields       []Field  // besides area,ts
	Colors       []string // fg, effects, bg

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

func LoadRuleEngine(fn string) (*RuleEngine, error) {
	cf, err := conf.Load(fn)
	if err != nil {
		return nil, err
	}

	this := new(RuleEngine)
	this.Conf = cf
	this.Workers = make([]ConfWorker, 0)
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
		parser.AbnormalPercent = this.Float(keyPrefix+"abnormal_percent", 1.5)
		parser.AbnormalBase = this.Int(keyPrefix+"abnormal_base", 10)

		// fields
		fields := this.List(keyPrefix+"fields", nil)
		if fields != nil {
			for j := 0; j < len(fields); j++ {
				prefix := fmt.Sprintf("%s[%d].", keyPrefix+"fields", j)
				field := Field{}
				field.Name = this.String(prefix+"name", "")
				field.Type = this.String(prefix+"type", "string")
				field.Contain = this.String(prefix+"contain", "")
				field.Sink = this.Int(prefix+"sink", 3)
				field.Ignores = this.StringList(prefix+"ignores", nil)
				field.Filters = this.StringList(prefix+"filters", nil)
				field.Regex = this.StringList(prefix+"regex", nil)
				parser.Fields = append(parser.Fields, field)

				if field.Contain != "" {
					// validator only, will never sink to db or indexer
					field.Sink = 0
				}
			}
		}

		this.Parsers = append(this.Parsers, parser)
	}

	// workers section
	workers := this.List("workers", nil)
	for i := 0; i < len(workers); i++ {
		keyPrefix := fmt.Sprintf("workers[%d].", i)
		worker := ConfWorker{}
		worker.Enabled = this.Bool(keyPrefix+"enabled", true)
		worker.Dsn = this.String(keyPrefix+"dsn", "file://")
		worker.TailGlob = this.String(keyPrefix+"tail_glob", "")
		worker.HistoryGlob = this.String(keyPrefix+"history_glob", "")
		worker.Parsers = this.StringList(keyPrefix+"parsers", nil)

		this.Workers = append(this.Workers, worker)
	}

	// validation
	if this.hasDupParsers() {
		return nil, errors.New("has dup parsers")
	}

	return this, nil
}

func (this *RuleEngine) IsParserApplied(parser string) bool {
	for _, g := range this.Workers {
		for _, p := range g.Parsers {
			if p == parser {
				return true
			}
		}
	}

	return false
}

// Dup parser id
func (this *RuleEngine) hasDupParsers() bool {
	parsers := make(map[string]bool)
	for _, p := range this.Parsers {
		if _, present := parsers[p.Id]; present {
			return true
		}

		parsers[p.Id] = true
	}

	return false
}

func (this *RuleEngine) CountOfWorkers() (c int) {
	for _, g := range this.Workers {
		if g.Enabled {
			c += 1
		}
	}

	return
}

func (this *RuleEngine) ParserById(id string) *ConfParser {
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

func (this *ConfParser) FieldByName(name string) (field Field, err error) {
	for _, lk := range this.Fields {
		if lk.Name == name {
			return lk, nil
		}
	}

	return Field{Name: ""}, errors.New("field: " + name + ": not found")
}

func (this *ConfWorker) HasParser(parserId string) bool {
	for _, pid := range this.Parsers {
		if pid == parserId {
			return true
		}
	}

	return false
}

func (this *ConfWorker) Scheme() string {
	u, err := url.Parse(this.TailGlob)
	if err != nil {
		panic(err)
	}

	return u.Scheme
}
