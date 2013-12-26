package main

import (
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/goserver"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func init() {
	goserver.RegisterSignalHandler(syscall.SIGINT, handleInterrupt)
	goserver.RegisterSignalHandler(syscall.SIGTERM, handleInterrupt)
	goserver.RegisterSignalHandler(syscall.SIGUSR2, func(sig os.Signal) {
		// TODO reload rule engine
	})
}

func handleInterrupt(sig os.Signal) {
	logger.Printf("got signal %s\n", strings.ToUpper(sig.String()))

	shutdown()
}
