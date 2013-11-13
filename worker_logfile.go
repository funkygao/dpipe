package main

import (
	"fmt"
	"github.com/funkygao/alser/config"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/tail"
	"os"
	"sync"
)

type LogfileWorker struct {
	Worker

	tailConf tail.Config
}

func newLogfileWorker(id int,
	dataSource string, conf config.ConfGuard, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int, chAlarm chan<- parser.Alarm) Runnable {
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
		chLines: chLines, chAlarm: chAlarm}
	this.tailConf = tailConfig

	return this
}

func (this LogfileWorker) String() string {
	return fmt.Sprintf("log.worker-%d[%s]", this.id, this.dataSource)
}

func (this *LogfileWorker) Run() {
	defer this.Done()

	t, err := tail.TailFile(this.dataSource, this.tailConf)
	if err != nil {
		panic(err)
	}
	defer t.Stop()

	if options.verbose {
		logger.Printf("%s started\n", *this)
	}

	for line := range t.Lines {
		// a valid line scanned
		this.chLines <- 1

		// feed the parsers one by one
		for _, parserId := range this.conf.Parsers {
			parser.Dispatch(parserId, line.Text)
		}
	}

	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}
}
