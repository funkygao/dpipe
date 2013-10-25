package main

import (
	"os"
	"runtime/pprof"
	"syscall"
)

func cleanup() {
	if options.lock {
		syscall.Unlink(LOCKFILE) // cleanup lock file
	}

	if options.pprof != "" {
		pprof.StopCPUProfile()
	}
}

func shutdown() {
	cleanup()
	logger.Println("terminated")
	os.Exit(0)
}
