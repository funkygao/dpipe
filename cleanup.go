package main

import (
	"github.com/funkygao/goserver"
	"os"
	"runtime/pprof"
)

func cleanup() {
	if options.lock {
		goserver.UnlockInstance(LOCKFILE)
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
