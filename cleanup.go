package main

import (
	"os"
	"runtime/pprof"
	"syscall"
)

func cleanup() {
	syscall.Unlink(lockfile) // cleanup lock file

	if options.pprof != "" {
		pprof.StopCPUProfile()
	}
}

func shutdown() {
	cleanup()
	os.Exit(0)
}
