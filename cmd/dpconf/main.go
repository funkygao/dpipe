package main

import (
	"flag"
	"fmt"
	conf "github.com/funkygao/jsconf"
	"os"
	"time"
)

const (
	IDENT    = "ident"
	MATCH    = "match"
	NAME     = "name"
	DISABLED = "disabled"
)

type identEntry struct {
	disabled bool
	matches  []string
}

var (
	graph map[string]*identEntry
)

func main() {
	var fn string
	flag.StringVar(&fn, "c", "", "config file path")
	flag.Parse()

	cf, err := conf.Load(fn)
	if err != nil {
		fmt.Printf("Invalid config file[%s]: %v\n", fn, err)
		os.Exit(1)
	}

	graph = make(map[string]*identEntry)

	// only visualize the plugins section
	for i := 0; i < len(cf.List("plugins", nil)); i++ {
		section, err := cf.Section(fmt.Sprintf("plugins[%d]", i))
		if err != nil {
			panic(err)
		}

		handleSection(section)
	}

	showGraph()
}

// recursively
func handleSection(section *conf.Conf) {
	ident := section.String(IDENT, "")
	if ident != "" {
		if _, present := graph[ident]; present {
			fmt.Printf("ident[%s]duplicated\n", ident)
			os.Exit(1)
		}

		ie := &identEntry{}
		ie.matches = make([]string, 0, 10)
		ie.disabled = section.Bool(DISABLED, false)

		graph[ident] = ie
	}

	matches := section.StringList(MATCH, nil)
	if matches != nil {
		pluginName := section.String(NAME, "")
		if pluginName == "" {
			fmt.Printf("plugin match %v has no 'name' key\n", matches)
			os.Exit(1)
		}

		for _, id := range matches {
			if _, present := graph[id]; present {
				graph[id].matches = append(graph[id].matches, pluginName)
			} else {
				fmt.Printf("%15s -> %s\n", id, pluginName)
			}
		}
	}

	sub := section.Interface("", nil).(map[string]interface{})
	if sub == nil {
		return
	}

	for k, v := range sub {
		if x, ok := v.([]interface{}); ok {
			switch x[0].(type) {
			case string, float64, int, bool, time.Time:
				// this section will never find 'ident'
				continue
			}

			for i := 0; i < len(section.List(k, nil)); i++ {
				key := fmt.Sprintf("%s[%d]", k, i)
				sec, err := section.Section(key)
				if err != nil {
					continue
				}

				handleSection(sec)
			}
		}
	}
}

func showGraph() {
	fmt.Println()
	for ident, ie := range graph {
		var flag = "△"
		if ie.disabled {
			flag = "▼"
		}

		fmt.Printf("%15s[%s] -> %v\n", ident, flag, ie.matches)
	}
}
