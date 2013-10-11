/*
          main
           |<-------------------------------
           |                                |
           | goN(wait group)                |
           V                                |
     -----------------------                |
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
	parser "github.com/funkygao/alsparser"
	"path/filepath"
	"sync"
	"time"
)

func guard(jsonConfig jsonConfig) {
	parser.SetLogger(logger)
	parser.SetVerbose(options.verbose)
	parser.SetDebug(options.debug)
	parser.SetDryRun(options.dryrun)

	var lines int = 0
	var workerN int = 0
	var wg = new(sync.WaitGroup)
	chLines := make(chan int)
	chAlarm := make(chan parser.Alarm, 1000)

	// create all parsers at once
	parser.NewParsers(jsonConfig.parsers(), chAlarm)

	// loop through the whole config
	for _, item := range jsonConfig {
		paths, err := filepath.Glob(item.Pattern)
		if err != nil {
			panic(err)
		}

		for _, logfile := range paths {
			workerN++
			wg.Add(1)

			// each logfile is a dedicated goroutine worker
			go run_worker(logfile, item, wg, chLines)
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

	parser.StopAll()

	logger.Println("all lines scaned:", lines)
}
