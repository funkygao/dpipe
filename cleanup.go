package main

import (
	"os"
	"runtime/pprof"
)

func cleanup() {
	if options.lock {
		unlockInstance()
	}

	if options.cpuprof != "" {
		pprof.StopCPUProfile()
	}

	if ticker != nil {
		ticker.Stop()
	}
}

func shutdown() {
	cleanup()

	logger.Println("terminated")
	os.Exit(0)
}
