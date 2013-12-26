package main

import (
	"github.com/funkygao/golib"
	"os"
	"runtime/pprof"
)

func cleanup() {
	if options.lock {
		golib.UnlockInstance(LOCKFILE)
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
