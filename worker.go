package main

import (
	"fmt"
	"github.com/funkygao/alser/config"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/tail"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type worker struct {
	id       int
	logfile  string // a single file
	conf     config.ConfGuard
	tailMode bool
	tailConf tail.Config
	*sync.Mutex
	wg      *sync.WaitGroup
	chLines chan<- int
	chAlarm chan<- parser.Alarm
}

func invokeWorkers(wg *sync.WaitGroup, workersCanWait chan<- bool, conf *config.Config, chLines chan<- int, chAlarm chan<- parser.Alarm) {
	allWorkers = make(map[string]bool)
	workersCanWaitOnce := new(sync.Once)
	mutex := new(sync.Mutex)

	// main loop to watch for newly emerging logfiles
	for {
		for _, g := range conf.Guards {
			if options.parser != "" && !g.HasParser(options.parser) {
				// only one parser applied
				continue
			}

			var pattern string
			if options.tailmode {
				pattern = g.TailLogGlob
			} else {
				pattern = g.HistoryLogGlob
			}

			logfiles, err := filepath.Glob(pattern)
			if err != nil {
				panic(err)
			}

			for _, logfile := range logfiles {
				if _, present := allWorkers[logfile]; present {
					// this logfile is already being guarded
					continue
				}

				wg.Add(1)
				allWorkers[logfile] = true

				// each logfile is a dedicated goroutine worker
				worker := newWorker(len(allWorkers), logfile, g, options.tailmode, wg, mutex, chLines, chAlarm)
				go worker.run()
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
}

func newWorker(id int, logfile string, conf config.ConfGuard, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int, chAlarm chan<- parser.Alarm) worker {
	this := worker{id: id, logfile: logfile, conf: conf, tailMode: tailMode,
		wg: wg, Mutex: mutex,
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

func (this worker) String() string {
	return fmt.Sprintf("worker-%d[%s]", this.id, this.logfile)
}

func (this *worker) run() {
	defer func() {
		this.wg.Done()

		this.Lock()
		delete(allWorkers, this.logfile) // map is not goroutine safe
		this.Unlock()
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

		// feed the parsers one by one
		for _, parserId := range this.conf.Parsers {
			parser.Dispatch(parserId, line.Text)
		}
	}

	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}
}
