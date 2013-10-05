package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

func init() {
	
}

func initialize(option *Option, err error) {
	if option.showversion {
		fmt.Fprintf(os.Stderr, "%s %s %s %s\n", "alser", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	options, err := ParseFlags()
	initialize(options, err)

	start := time.Now()

}
