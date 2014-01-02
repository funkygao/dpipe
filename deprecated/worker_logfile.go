package main

import (
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/alser/rule"
	"github.com/funkygao/tail"
	"os"
	"sync"
)

type LogfileWorker struct {
	Worker

	tailConf tail.Config
}

func newLogfileWorker(id int,
	dataSource string, conf rule.ConfWorker, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int) Runnable {
	var tailConfig tail.Config
	if tailMode {
		tailConfig = tail.Config{
			Follow:   true, // tail -f
			Poll:     true, // Poll for file changes instead of using inotify
			ReOpen:   true, // tail -F
			Location: &tail.SeekInfo{Offset: int64(0), Whence: os.SEEK_END},
			//MustExist: false,
		}
	}

	this := new(LogfileWorker)
	this.Worker = Worker{id: id,
		dataSource: dataSource, conf: conf, tailMode: tailMode,
		wg: wg, Mutex: mutex,
		chLines: chLines}
	this.tailConf = tailConfig

	return this
}

func (this *LogfileWorker) Run() {
	defer this.Done()

	t, err := tail.TailFile(this.dataSource, this.tailConf)
	if err != nil {
		panic(err)
	}
	defer t.Stop()

	if options.verbose {
		logger.Printf("%s started", *this)
	}

	for line := range t.Lines {
		// a valid line scanned
		this.chLines <- 1

		// feed the parsers one by one
		for _, parserId := range this.conf.Parsers {
			if options.debug {
				logger.Printf("%s got line: %s\n", *this, line.Text)
			}

			parser.Dispatch(parserId, line.Text)
		}
	}

	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}
}
