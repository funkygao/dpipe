package main

import (
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/locking"
	"os"
	"runtime/pprof"
	"time"
)

func cleanup() {
	if options.lockfile != "" {
		locking.UnlockInstance(options.lockfile)
	}

	if options.cpuprof != "" {
		pprof.StopCPUProfile()
	}

	if options.memprof != "" {
		f, err := os.Create(options.memprof)
		if err != nil {
			panic(err)
		}

		globals.Printf("MEM profiler %s enabled\n", options.memprof)
		pprof.WriteHeapProfile(f)
		f.Close()
	}
}

func shutdown() {
	cleanup()

	globals.Printf("Terminated after %s.\n", time.Since(engine.Globals().StartedAt))
	os.Exit(0)
}
