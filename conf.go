package main

import (
	"os"
	"encoding/json"
	"fmt"
)

type Record struct {
	Name string
	Parsers []string
	Pattern string
}

type Config struct {
	Records []Record
}

func loadConfig(filename string) (config *Config)  {
	file, e := os.Open(filename)
	if e != nil {
		panic(e)
	}

	decoder := json.NewDecoder(file)
	config = &Config{}
	if e := decoder.Decode(&config); e != nil {
		panic(e)
	}

	if options.verbose {
		logger.Printf("%#v\n", *config)
	}

	return
}
