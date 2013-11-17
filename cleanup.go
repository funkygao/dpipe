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
