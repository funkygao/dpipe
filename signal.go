package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type signalHandler func(os.Signal)

var signals = struct {
	sync.Mutex // map in go is not thread/goroutine safe
	handlers   map[os.Signal]signalHandler
	c          chan os.Signal
}{
	handlers: make(map[os.Signal]signalHandler),
	c:        make(chan os.Signal, 20),
}

func trapSignals() {
	for sig := range signals.c {
		signals.Lock()
		handler := signals.handlers[sig]
		signals.Unlock()
		if handler != nil {
			handler(sig)
		}
	}
}

func registerSignalHandler(sig os.Signal, handler signalHandler) {
	signals.Lock()
	defer signals.Unlock()

	if _, alreadyExists := signals.handlers[sig]; !alreadyExists {
		signals.handlers[sig] = handler
		signal.Notify(signals.c, sig)
	}
}

func handleIgnore(sig os.Signal) {}

func handleInterrupt(sig os.Signal) {
	logger.Printf("got signal %s\n", strings.ToUpper(sig.String()))
	logger.Println("terminated")
	shutdown()
}

func setupSignals() {
	go trapSignals()

	registerSignalHandler(syscall.SIGINT, handleInterrupt)
}
