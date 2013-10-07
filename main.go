package main

import (
	"fmt"
	"os"
)

func init() {
	options = parseFlags()
	options.validate()
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintln(os.Stderr, e)
		}
	}()

	logger = newLogger(options)
	logger.Println("started")

	loadConfig(options.config)

}
