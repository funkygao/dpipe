package main

import (
	"os"
	"encoding/json"
)

type jsonItem struct {
	Name string `json:"name"` // must be exported, else json can't unmarshal
	Pattern string `json:"pattern"` // may not be same order as json file
	Parsers []string
}

type jsonConfig []jsonItem

func loadConfig(filename string) (config jsonConfig)  {
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
