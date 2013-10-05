package main

import (
	"fmt"
	"os"
	"runtime"
	"log"
	//"time"
)

func init() {
	
}

func initialize(option *Option, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if option.showversion {
		fmt.Fprintf(os.Stderr, "%s %s %s %s\n", "alser", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

func main() {
	options, err := ParseFlags()
	initialize(options, err)

	//start := time.Now()
	var logger *log.Logger = newLogger(options)
	logger.Println("started")

}
