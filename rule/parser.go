package rule

import (
	"errors"
	"fmt"
)

type ConfParser struct {
	Id           string // required
	Class        string // required
	Title        string // optional, defaults to Id
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

func (this *ConfParser) validate() {
	if this.Id == "" || this.Class == "" {
		panic("parser Id/Class can't be empty")
	}

	if this.Title == "" {
		this.Title = this.Id
	}
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
