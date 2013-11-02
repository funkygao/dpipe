package main

import (
	"github.com/funkygao/alser/config"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/gofmt"
	//"path/filepath"
	"runtime"
	"time"
)

func runTicker(ticker *time.Ticker, lines *int) {
	ms := new(runtime.MemStats)
	for _ = range ticker.C {
		runtime.ReadMemStats(ms)
		logger.Printf("ver:%s, goroutine:%d, mem:%s, workers:%d parsers:%d lines:%d, elapsed:%s\n",
			BuildID,
			runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc),
			len(allWorkers), parser.ParsersCount(), *lines, time.Since(startTime))
	}
}

func runAlarmCollector(ch <-chan parser.Alarm) {
	// we don't know when to send alarm, we just send alarm one by one
	// alarm can span several lines
	// it's parsers' responsibility for flow control such as backoff
	for alarm := range ch {
		// TODO send email
		logger.Println(alarm)
	}
}

func notifyUnGuardedLogs(conf *config.Config) {
	// what logs are still out of our guard?

}
