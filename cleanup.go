package main

import (
	"github.com/funkygao/golib/locking"
	"os"
	"runtime/pprof"
)

func cleanup() {
	if options.lockfile != "" {
		locking.UnlockInstance(options.lockfile)
	}

	if options.cpuprof != "" {
		pprof.StopCPUProfile()
	}
}

func shutdown() {
	cleanup()

	globals.Logger.Println("terminated")
	os.Exit(0)
}
