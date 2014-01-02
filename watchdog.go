package main

import (
	"github.com/funkygao/gofmt"
	"runtime"
	"time"
)

func runTicker(ticker *time.Ticker) {
	startTime := time.Now()
	ms := new(runtime.MemStats)
	for _ = range ticker.C {
		runtime.ReadMemStats(ms)

		globals.Logger.Printf("ver:%s, goroutine:%d, mem:%s, elapsed:%s\n",
			BuildID,
			runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc),
			time.Since(startTime))
	}
}
