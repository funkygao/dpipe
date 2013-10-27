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

	var workersWg = new(sync.WaitGroup)
	chLines := make(chan int)
	wgCanWait := make(chan bool) // in case of wg.Add/Wait race condition
	go prepareWorkers(workersWg, wgCanWait, jsonConfig, chLines)

	// wait for all workers finish
	go func() {
		<-wgCanWait
		workersWg.Wait()

		logger.Println("all workers finished")

		close(chLines)
		close(chAlarm)
	}()

	// after all workers finished, collect how many lines scanned
	for l := range chLines {
		lines += l
	}

	parser.StopAll()

	logger.Printf("%d lines scanned, %s elapsed\n", lines, time.Since(startTime))
}

func prepareWorkers(wg *sync.WaitGroup, wgCanWait chan<- bool, jsonConfig jsonConfig, chLines chan<- int) {
	guardedFiles := make(map[string]bool)
	wgCanWaitSent := false

	for {
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
				if _, present := guardedFiles[logfile]; present {
					// this logfile is already beting tailed
					continue
				}

				guardedFiles[logfile] = true
				wg.Add(1)

				// each logfile is a dedicated goroutine worker
				go runWorker(logfile, item, wg, chLines)
				if options.verbose {
					logger.Printf("worker[%s] started\n", logfile)
				}
			}
		}

		if !wgCanWaitSent {
			wgCanWait <- true
			wgCanWaitSent = true
		}

		if !options.tailmode {
			break
		} else {
			<-time.After(time.Second * 2)
		}
	}

	if options.parser != "" {
		logger.Printf("only parser %s running\n", options.parser)
	}
	logger.Println(len(guardedFiles), "workers started")
}
