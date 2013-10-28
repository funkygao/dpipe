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
	if options.tick > 0 { // ticker for reporting workers progress
		ticker := time.NewTicker(time.Second * time.Duration(options.tick))
		go runTicker(ticker, &lines)
	}

	chAlarm := make(chan parser.Alarm, 1000) // collect alarms from all parsers
	go runAlarmCollector(chAlarm)            // unified alarm handling

	var workersWg = new(sync.WaitGroup)
	chLines := make(chan int)
	workersCanWait := make(chan bool) // in case of wg.Add/Wait race condition
	go prepareWorkers(workersWg, workersCanWait, jsonConfig, chLines, chAlarm)

	// wait for all workers finish
	go func() {
		<-workersCanWait
		workersWg.Wait()

		close(chLines)
		close(chAlarm)
	}()

	// after all workers finished, collect how many lines scanned
	for l := range chLines {
		lines += l
	}

	if options.verbose {
		logger.Println("all lines are fed to parsers, stopping all parsers...")
	}
	parser.StopAll()

	if options.verbose {
		logger.Println("awaiting all parsers...")
	}
	parser.WaitAll()

	logger.Printf("%d lines scanned, %s elapsed\n", lines, time.Since(startTime))
}

func prepareWorkers(wg *sync.WaitGroup, workersCanWait chan<- bool, jsonConfig jsonConfig, chLines chan<- int, chAlarm chan<- parser.Alarm) {
	allWorkers = make(map[string]bool)
	workersCanWaitOnce := new(sync.Once)

	// main loop to watch for newly emerging logfiles
	for {
		for _, item := range jsonConfig {
			if options.parser != "" && !item.hasParser(options.parser) {
				// only one parser applied
				continue
			}

			logfiles, err := filepath.Glob(item.Pattern)
			if err != nil {
				panic(err)
			}

			if options.debug {
				logger.Printf("glob: %s, got: %+v\n", item.Pattern, logfiles)
			}

			for _, logfile := range logfiles {
				if _, present := allWorkers[logfile]; present {
					// this logfile is already being tailed
					continue
				}

				allWorkers[logfile] = true
				wg.Add(1)

				// each logfile is a dedicated goroutine worker
				go runWorker(logfile, item, wg, chLines, chAlarm)
				if options.verbose {
					logger.Printf("worker-%d[%s] started\n", logfile, len(allWorkers))
				}
			}
		}

		workersCanWaitOnce.Do(func() {
			workersCanWait <- true
		})

		if !options.tailmode {
			break
		} else {
			<-time.After(time.Second * 2)
		}
	}

	if options.parser != "" {
		logger.Printf("only parser %s running\n", options.parser)
	}
	logger.Println(len(allWorkers), "workers started")
}
