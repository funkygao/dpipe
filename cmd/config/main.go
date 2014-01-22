package main

import (
	"flag"
	"fmt"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/pretty"
)

const (
	IDENT = "ident"
)

var (
	graph map[string][]string
)

func main() {
	var fn string
	flag.StringVar(&fn, "c", "", "config file path")
	flag.Parse()

	cf, err := conf.Load(fn)
	if err != nil {
		panic(err)
	}

	graph = make(map[string][]string)

	for i := 0; i < len(cf.List("plugins", nil)); i++ {
		section, err := cf.Section(fmt.Sprintf("plugins[%d]", i))
		if err != nil {
			panic(err)
		}

		handleSection(section)
	}

	showGraph()
}

func handleSection(section *conf.Conf) {
	ident := section.String(IDENT, "")
	if ident != "" {
		graph[ident] = make([]string, 0, 10)
	}

	sub := section.Interface("", nil).(map[string]interface{})
	if sub != nil {
		for k, v := range sub {
			//pretty.Printf("%s => %v\n", k, v)

			if _, ok := v.([]interface{}); ok {
				pretty.Printf("%s => %v\n", k, v)
				continue

				if _, strList := v.([]string); strList {
					fmt.Println("haha", k)
					continue
				}
				for i := 0; i < len(section.List(k, nil)); i++ {
					key := fmt.Sprintf("%s[%d]", k, i)
					fmt.Println(key)
					sec, err := section.Section(key)
					if err != nil {
						continue
					}

					handleSection(sec)
				}

			}
		}
	}
}

func showGraph() {
	pretty.Printf("%# v", graph)
}
