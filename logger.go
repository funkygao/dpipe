package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func newLogger(option *Option) *log.Logger {
	var logWriter io.Writer = os.Stdout // default log writer
	var err error
	if option.logfile != "" {
		logWriter, err = os.OpenFile(option.logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
	}

	prefix := fmt.Sprintf("[%d]", os.Getpid()) // prefix with pid
	if option.debug {
		return log.New(logWriter, prefix, LOG_OPTIONS_DEBUG)
	}

	return log.New(logWriter, prefix, LOG_OPTIONS)
}
