package main

import (
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/gofmt"
	"runtime"
)

func runTicker(lines *int) {
	ms := new(runtime.MemStats)
	for _ = range ticker.C {
		runtime.ReadMemStats(ms)
		logger.Printf("goroutine: %d, mem: %s, lines: %d\n",
			runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc), *lines)
	}
}

func runAlarmCollector(ch <-chan parser.Alarm) {
	alarmLogger := newAlarmLogger()
	for alarm := range ch {
		alarmLogger.Println(alarm)
	}
}
