package main

import (
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/tail"
	"os"
	"sync"
)

// Each single log file is a worker
// Workers share some singleton parsers
func runWorker(logfile string, conf jsonItem, wg *sync.WaitGroup, chLines chan<- int) {
	defer func() {
		wg.Done()
		delete(guardedFiles, logfile)
	}()

	var tailConfig tail.Config
	if options.tailmode {
		tailConfig = tail.Config{
			Follow:   true, // tail -f
			Poll:     true, // Poll for file changes instead of using inotify
			ReOpen:   true, // tail -F
			Location: &tail.SeekInfo{Offset: int64(0), Whence: os.SEEK_END},
			//MustExist: false,
		}
	}

	t, err := tail.TailFile(logfile, tailConfig)
	if err != nil {
		panic(err)
	}

	defer t.Stop()

	for line := range t.Lines {
		// a valid line scanned
		chLines <- 1

		for _, p := range conf.Parsers {
			parser.Dispatch(p, line.Text)
		}
	}

	if options.verbose {
		logger.Println(logfile, "finished")
	}
}
