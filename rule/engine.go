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
	"fmt"
	conf "github.com/daviddengcn/go-ljson-conf"
)

type RuleEngine struct {
	*conf.Conf

	Workers []ConfWorker
	Parsers map[string]*ConfParser
}

func LoadRuleEngine(fn string) (*RuleEngine, error) {
	cf, err := conf.Load(fn)
	if err != nil {
		return nil, err
	}

	this := new(RuleEngine)
	this.Conf = cf
	this.Workers = make([]ConfWorker, 0)
	this.Parsers = make(map[string]*ConfParser)

	// parsers section
	parsers := this.List("parsers", nil)
	for i := 0; i < len(parsers); i++ {
		keyPrefix := fmt.Sprintf("parsers[%d].", i)
		parser := new(ConfParser)
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

		if _, present := this.Parsers[parser.Id]; present {
			panic("parser with id:" + parser.Id + " already exists")
		} else {
			this.Parsers[parser.Id] = parser
		}

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

	return this, nil
}

func (this *RuleEngine) IsParserApplied(parserId string) bool {
	for _, g := range this.Workers {
		for _, pid := range g.Parsers {
			if pid == parserId {
				return true
			}
		}
	}

	return false
}

func (this *RuleEngine) WorkersCount() (c int) {
	for _, g := range this.Workers {
		if g.Enabled {
			c += 1
		}
	}

	return
}

func (this *RuleEngine) ParserById(pid string) *ConfParser {
	p, present := this.Parsers[pid]
	if !present {
		return nil
	}

	return p
}

func (this *RuleEngine) DiscardParsersExcept(pid string) {
	for id, _ := range this.Parsers {
		if id != pid {
			delete(this.Parsers, id)
		}
	}
}
