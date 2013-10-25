package main

import (
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/gofmt"
	"runtime"
	"time"
)

func runTicker(lines *int) {
	ms := new(runtime.MemStats)
	for _ = range ticker.C {
		runtime.ReadMemStats(ms)
		logger.Printf("goroutine: %d, mem: %s, lines: %d, elapsed: %s\n",
			runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc),
			*lines, time.Since(startTime))
	}
}

func runAlarmCollector(ch <-chan parser.Alarm) {
	for alarm := range ch {
		logger.Println(alarm)
	}
}
