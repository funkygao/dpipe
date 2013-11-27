package main

import (
	"fmt"
	"github.com/mattbaird/elastigo/core"
)

func search(index, typ string, word string) {
	out, err := core.SearchRequest(true, index, typ, word, "", 0)
	if err != nil {
		panic(err)
	}

	for i := 0; i < out.Hits.Len(); i++ {
		fmt.Println(string(out.Hits.Hits[i].Source))
	}
}
