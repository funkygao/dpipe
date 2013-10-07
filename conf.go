package main

import (
	"os"
	"encoding/json"
)

type Record struct {
	Name string `json:"name"`
	Parsers []string
	Pattern string
}

type Config []Record

func loadConfig(filename string) (config *Config)  {
	file, e := os.Open(filename)
	if e != nil {
		panic(e)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if e := decoder.Decode(&config); e != nil {
		panic(e)
	}

	return
}
