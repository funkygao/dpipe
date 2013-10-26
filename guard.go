/*
          main
           |
           |<-------------------------------
           |                                |
           | goN(wait group)           -----------------
           V                          | alarm collector |
     -----------------------           -----------------
    |       |       |       |               |
   log1    log2    ...     logN             |
    |       |       |       |               | alarm
     -----------------------                | chan
           |                                |
           | feed lines                     |
           V                                |
     -----------------------                ^
    |       |       |       |               |
  parser1 parser2  ...   parserM            |
    |       |       |       |               |
     -----------------------                |
           |                                |
            ------------------->------------

*/
package main

import (
	"github.com/funkygao/alser/parser"
	"path/filepath"
	"sync"
	"time"
)

func guard(jsonConfig jsonConfig) {
	startTime = time.Now()

	// pass config to parsers
	parser.SetLogger(logger)
	parser.SetVerbose(options.verbose)
	parser.SetDebug(options.debug)
	parser.SetDryRun(options.dryrun)

	var lines int = 0
	if options.tick > 0 { // ticker watchdog for reporting workers progress
		ticker := time.NewTicker(time.Second * time.Duration(options.tick))
		go runTicker(ticker, &lines)
	}

	chAlarm := make(chan parser.Alarm, 1000) // collect alarms from all parsers
	go runAlarmCollector(chAlarm)            // unified alarm handling

	// create all parsers at once FIXME what if options.parser
	parser.NewParsers(jsonConfig.parsers(), chAlarm)

	var workerN int = 0
	var wg = new(sync.WaitGroup)
	chLines := make(chan int)
	for _, item := range jsonConfig {
		if options.parser != "" && !item.hasParser(options.parser) {
			// item parser skipped
			continue
		}

		logfiles, err := filepath.Glob(item.Pattern)
		if err != nil {
			panic(err)
		}

		for _, logfile := range logfiles {
			workerN++
			wg.Add(1)

			// each logfile is a dedicated goroutine worker
			go runWorker(logfile, item, wg, chLines)
		}
	}

	if options.parser != "" {
		logger.Printf("only parser %s running\n", options.parser)
	}
	logger.Println(workerN, "workers started")

	// wait for all workers finish
	go func() {
		wg.Wait()
		logger.Println("all", workerN, "workers finished")

		close(chLines)
		close(chAlarm)
	}()

	// after all workers finished, collect how many lines scanned
	for l := range chLines {
		lines += l
	}

	if options.parser != "" {
		// FIXME
		// wait parser's collectAlarm done
		<-time.After(time.Minute * 3)
	}

	parser.StopAll()

	logger.Printf("%d lines scanned, %s elapsed\n", lines, time.Since(startTime))
}
