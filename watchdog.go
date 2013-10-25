package main

import (
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/gofmt"
	"runtime"
	"time"
)

func runTicker(ticker *time.Ticker, lines *int) {
	ms := new(runtime.MemStats)
	for _ = range ticker.C {
		runtime.ReadMemStats(ms)
		logger.Printf("goroutine: %d, mem: %s, lines: %d, elapsed: %s\n",
			runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc),
			*lines, time.Since(startTime))
	}
}

func runAlarmCollector(ch <-chan parser.Alarm) {
	// we don't when to send alarm
	// it's parsers' responsibility for flow control such as backoff
	for alarm := range ch {
		// TODO send email
		logger.Println(alarm)
	}
}
