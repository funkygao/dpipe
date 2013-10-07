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
	logger.Println("start...")

	jsonConfig := loadConfig(options.config)
	logger.Printf("json config has %d items\n", len(jsonConfig))
	if options.verbose {
		for i, item := range jsonConfig {
			logger.Printf("[%2d] %+v\n", i, item)
		}
	}

}
