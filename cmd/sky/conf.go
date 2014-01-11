package main

import (
	"errors"
	"fmt"
	conf "github.com/funkygao/jsconf"
)

var (
	allTables = make(map[string]table)
)

type property struct {
	name      string
	transient bool
	dataType  string // integer, string, float, boolean, factor
}

type table struct {
	name       string
	properties []property
}

func (this *table) init(config *conf.Conf) error {
	this.properties = make([]property, 0, 10)
	this.name = config.String("name", "")
	if this.name == "" {
		return errors.New("empty table name")
	}

	for i := 0; i < len(config.List("props", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("props[%d]", i))
		if err != nil {
			return err
		}

		p := property{}
		p.name = section.String("name", "")
		p.transient = section.Bool("transient", true)
		p.dataType = section.String("type", "")
		this.properties = append(this.properties, p)
	}

	return nil
}

func loadConfig(fn string) error {
	cf, err := conf.Load(fn)
	if err != nil {
		return err
	}

	for i := 0; i < len(cf.List("tables", nil)); i++ {
		section, err := cf.Section(fmt.Sprintf("tables[%d]", i))
		if err != nil {
			return err
		}

		t := table{}
		if err = t.init(section); err != nil {
			return err
		}
		allTables[t.name] = t
	}

	return nil
}
