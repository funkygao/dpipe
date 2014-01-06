package main

import (
	"flag"
	"github.com/funkygao/als"
	"os"
)

func main() {
	var gobfile string
	flag.StringVar(&gobfile, "f", "", "gob filename")
	flag.Parse()

	if gobfile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	chkpoint := als.NewFileCheckpoint(gobfile)
	chkpoint.PrettyPrint()
}
