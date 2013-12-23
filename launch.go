package main

import (
	"fmt"
	sqldb "github.com/funkygao/alser/db"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/alser/rule"
	"path/filepath"
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
	if options.logfile != "" || options.daemon {
		parser.SetBackground(true)
	}

	var lines int = 0
	if options.tick > 0 { // ticker for reporting workers progress
		ticker = time.NewTicker(time.Second * time.Duration(options.tick))
		go runTicker(ticker, &lines)
	}

	chAlarm := make(chan parser.Alarm, 1000) // collect alarms from all parsers
	go runAlarmCollector(chAlarm)            // unified alarm handling

	go notifyUnGuardedLogs(conf)

	parser.InitParsers(options.parser, conf, chAlarm)

	var workersWg = new(sync.WaitGroup)
	chLines := make(chan int)         // how many line have been scanned till now
	workersCanWait := make(chan bool) // in case of wg.Add/Wait race condition
	go invokeWorkers(conf, workersWg, workersCanWait, chLines, chAlarm)

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

func guardDataSources(guard config.ConfGuard) []string {
	if guard.Type == config.DATASOURCE_FILE || guard.Type == config.DATASOURCE_SYS {
		var pattern string
		if options.tailmode {
			pattern = guard.TailLogGlob
		} else {
			pattern = guard.HistoryLogGlob
		}

		logfiles, err := filepath.Glob(pattern)
		if err != nil {
			panic(err)
		}

		if options.debug {
			logger.Printf("pattern:%s -> %+v\n", pattern, logfiles)
		}

		return logfiles
	} else if guard.Type == config.DATASOURCE_DB {
		tables := make([]string, 0)
		db := sqldb.NewSqlDb(sqldb.DRIVER_MYSQL, FLASHLOG_DSN, logger)
		rows := db.Query(fmt.Sprintf("SHOW TABLES LIKE '%s'", guard.Tables))
		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err != nil {
				panic(err)
			}
			tables = append(tables, table)
		}

		db.Close()

		return tables
	} else {
		panic("unkown guards data source: " + guard.Type)
	}

	return nil
}
