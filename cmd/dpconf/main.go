package main

import (
	"flag"
	"fmt"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/pretty"
	"os"
)

const (
	IDENT = "ident"
	MATCH = "match"
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
		fmt.Printf("Invalid config file[%s]: %v", fn, err)
		os.Exit(1)
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

// recursive
func handleSection(section *conf.Conf) {
	ident := section.String(IDENT, "")
	if ident != "" {
		graph[ident] = make([]string, 0, 10)
	}

	sub := section.Interface("", nil).(map[string]interface{})
	if sub == nil {
		return
	}

	if sub != nil {
		for k, v := range sub {
			if x, ok := v.([]interface{}); ok {

				switch x[0].(type) {
				case string:
					continue
				case float64:
					continue
				}

				if _, ss := x[0].(string); ss {
					//continue
				}
				fmt.Printf("%s => %T\n", k, v)

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
