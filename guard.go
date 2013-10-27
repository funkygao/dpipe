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
	parser.SetDaemon(options.daemon)

	var lines int = 0
	if options.tick > 0 { // ticker watchdog for reporting workers progress
		ticker := time.NewTicker(time.Second * time.Duration(options.tick))
		go runTicker(ticker, &lines)
	}

	chAlarm := make(chan parser.Alarm, 1000) // collect alarms from all parsers
	go runAlarmCollector(chAlarm)            // unified alarm handling

	var workersWg = new(sync.WaitGroup)
	chLines := make(chan int)
	wgCanWait := make(chan bool) // in case of wg.Add/Wait race condition
	go prepareWorkers(workersWg, wgCanWait, jsonConfig, chLines, chAlarm)

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

	if options.verbose {
		logger.Println("stopping all parsers...")
	}
	parser.StopAll()

	if options.verbose {
		logger.Println("waiting all parsers...")
	}
	parser.WaitAll()

	logger.Printf("%d lines scanned, %s elapsed\n", lines, time.Since(startTime))
}

func prepareWorkers(wg *sync.WaitGroup, wgCanWait chan<- bool, jsonConfig jsonConfig, chLines chan<- int, chAlarm chan<- parser.Alarm) {
	guardedFiles = make(map[string]bool)
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

			if options.debug {
				logger.Printf("search pattern: %s, got: %+v\n", item.Pattern, logfiles)
			}

			for _, logfile := range logfiles {
				if _, present := guardedFiles[logfile]; present {
					// this logfile is already beting tailed
					continue
				}

				guardedFiles[logfile] = true
				wg.Add(1)

				// each logfile is a dedicated goroutine worker
				go runWorker(logfile, item, wg, chLines, chAlarm)
				if options.verbose {
					logger.Printf("worker[%s]-%d started\n", logfile, len(guardedFiles))
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
