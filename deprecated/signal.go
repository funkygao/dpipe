package main

import (
	"github.com/funkygao/golib/signal"
	"os"
	"strings"
	"sync"
	"syscall"
)

func init() {
	signal.RegisterSignalHandler(sig, handler)
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
