package main

import (
	"github.com/funkygao/alser/config"
	"github.com/funkygao/alser/parser"
	"sync"
	"time"
)

func guard(conf *config.Config) {
	startTime := time.Now()

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

	go notifyUnGuardedLogs(conf)

	parser.InitParsers(options.parser, conf, chAlarm)

	var workersWg = new(sync.WaitGroup)
	chLines := make(chan int)         // how many line have been scanned till now
	workersCanWait := make(chan bool) // in case of wg.Add/Wait race condition
	go invokeWorkers(workersWg, workersCanWait, conf, chLines, chAlarm)

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
		logger.Println("awaiting all parsers collecting alarms...")
	}
	parser.WaitAll()

	logger.Printf("%d lines scanned, %s elapsed\n", lines, time.Since(startTime))
}
