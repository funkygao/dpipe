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
	sync.Mutex // map in go is not thread safe
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

/*
syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT,
        syscall.SIGHUP, syscall.SIGSTOP, syscall.SIGQUIT,
*/
func registerSignalHandler(sig os.Signal, handler signalHandler) {
	signals.Lock()
	defer signals.Unlock()

	if _, alreadyExists := signals.handlers[sig]; !alreadyExists {
		signals.handlers[sig] = handler
		signal.Notify(signals.c, sig)
	}
}

func handlerIgnore(sig os.Signal) {}

func handlerInterrupt(sig os.Signal) {
	logger.Printf("%s signal recved\n", strings.ToUpper(sig.String()))
	logger.Println("terminated")
	shutdown()
}

func setupSignals() {
	go trapSignals()

	registerSignalHandler(syscall.SIGINT, handlerInterrupt)

}
