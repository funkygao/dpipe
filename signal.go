package main

import (
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/golib"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func init() {
	golib.RegisterSignalHandler(syscall.SIGINT, handleInterrupt)
	golib.RegisterSignalHandler(syscall.SIGTERM, handleInterrupt)
	golib.RegisterSignalHandler(syscall.SIGUSR2, func(sig os.Signal) {
		// TODO reload rule engine
	})
}

func handleInterrupt(sig os.Signal) {
	logger.Printf("got signal %s\n", strings.ToUpper(sig.String()))

	shutdown()
}
