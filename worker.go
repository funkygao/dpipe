package main

import (
	"fmt"
	"github.com/funkygao/alser/config"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/tail"
	"os"
	"sync"
)

type worker struct {
	id       int
	logfile  string // a single file
	conf     config.ConfGuard
	tailMode bool
	tailConf tail.Config
	wg       *sync.WaitGroup
	chLines  chan<- int
	chAlarm  chan<- parser.Alarm
}

func newWorker(id int, logfile string, conf config.ConfGuard, tailMode bool,
	wg *sync.WaitGroup,
	chLines chan<- int, chAlarm chan<- parser.Alarm) worker {
	this := worker{id: id, logfile: logfile, conf: conf, tailMode: tailMode,
		wg:      wg,
		chLines: chLines, chAlarm: chAlarm}

	var tailConfig tail.Config
	if this.tailMode {
		tailConfig = tail.Config{
			Follow:   true, // tail -f
			Poll:     true, // Poll for file changes instead of using inotify
			ReOpen:   true, // tail -F
			Location: &tail.SeekInfo{Offset: int64(0), Whence: os.SEEK_END},
			//MustExist: false,
		}
	}
	this.tailConf = tailConfig

	return this
}

func (this *worker) String() string {
	return fmt.Sprintf("worker-%d[%s]", this.id, this.logfile)
}

func (this *worker) run() {
	defer func() {
		this.wg.Done()
		delete(allWorkers, this.logfile) // FIXME not goroutine safe
	}()

	t, err := tail.TailFile(this.logfile, this.tailConf)
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

		for _, p := range this.conf.Parsers {
			parser.Dispatch(p, line.Text)
		}
	}

	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}
}
