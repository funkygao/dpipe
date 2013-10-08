/*
           main
           ^  |
     alarm |  | goN(wait group)
           |  V
     -----------------------
    |       |       |       |
   log1    log2    ...     logN
    |       |       |       |
  parsers parsers        parsers

*/
package main

import (
    //"github.com/funkygao/alser/parser"
	"./parser"
    "github.com/funkygao/gofmt"
    "path/filepath"
    "runtime"
    "sync"
    "time"
)

func guard(jsonConfig jsonConfig) {
    parser.SetLogger(logger)
    parser.SetVerbose(options.verbose)
    parser.SetDebug(options.debug)

    var lines int = 0
    var workerN int = 0
    var wg = new(sync.WaitGroup)
    chLines := make(chan int)
	chAlarm := make(chan parser.Alarm, 1000)
    for _, item := range jsonConfig {
        paths, err := filepath.Glob(item.Pattern)
        if err != nil {
            panic(err)
        }

        for _, logfile := range paths {
            workerN++
            wg.Add(1)
            go run_worker(logfile, item, wg, chLines, chAlarm)
        }
    }

    if options.tick > 0 {
        ticker = time.NewTicker(time.Second * time.Duration(options.tick))
        go runTicker(&lines)
    }

    logger.Println(workerN, "workers started")

	go runAlarmCollector(chAlarm)

    // wait for all workers finish
    go func() {
        wg.Wait()
        logger.Println("all", workerN, " workers finished")
        close(chLines)
		close(chAlarm)
    }()

    // collect how many lines scanned
    for l := range chLines {
        lines += l
    }

    logger.Println("all lines scaned:", lines)
}

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
		alarmLogger.Printf("%s,%v,%v\n", alarm.Area, alarm.Duration, alarm.Info)
	}
}
