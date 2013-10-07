package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

func init() {
	options = parseFlags()
	options.validate()
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			fmt.Fprintln(os.Stderr, e)
		}
	}()

	logger = newLogger(options)
	logger.Println("started")

	config := loadConfig(options.config)
	if options.verbose {
		logger.Printf("%#v\n", config)
	}

}
