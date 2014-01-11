package main

import (
	"flag"
	sky "github.com/funkygao/skyapi"
)

func main() {
	var (
		fn   string
		host string
		port int
	)
	flag.StringVar(&fn, "c", "tables.cf", "config filename")
	flag.StringVar(&host, "h", "localhost", "sky server address")
	flag.IntVar(&port, "p", 8585, "sky server port")
	flag.Parse()

	if err := loadConfig(fn); err != nil {
		panic(err)
	}

	client := sky.NewClient(host)
	client.Port = port
	if !client.Ping() {
		panic("server is down")
	}

	for _, t := range allTables {
		table, _ := client.GetTable(t.name)
		if table == nil {
			createTable(client, t)
		}
	}

}

func createTable(client *sky.Client, t table) {
	table := sky.NewTable(t.name, client)
	if err := client.CreateTable(table); err != nil {
		panic(err)
	}

	for _, p := range t.properties {
		table.CreateProperty(sky.NewProperty(p.name, p.transient, p.dataType))
	}
}
