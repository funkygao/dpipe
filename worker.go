package main

import (
	"github.com/funkygao/alser/config"
	"github.com/funkygao/alser/parser"
	"sync"
	"time"
)

type NewWorker func(id int, dataSource string,
	conf config.ConfGuard, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int, chAlarm chan<- parser.Alarm) Runnable

type Runnable interface {
	Run()
}

type Stringable interface {
	String() string
}

type Finishable interface {
	Done()
}

type Worker struct {
	Runnable
	Stringable
	Finishable

	id         int
	dataSource string // a single file or a single db table
	conf       config.ConfGuard
	tailMode   bool

	*sync.Mutex
	wg      *sync.WaitGroup
	chLines chan<- int
	chAlarm chan<- parser.Alarm
}

func (this Worker) String() string {
	return fmt.Sprintf("log.worker-%d[%s]", this.id, this.dataSource)
}

func (this *Worker) Done() {
	this.wg.Done()

	this.Lock()
	delete(allWorkers, this.dataSource) // map is not goroutine safe
	this.Unlock()
}

func invokeWorkers(conf *config.Config, wg *sync.WaitGroup, workersCanWait chan<- bool, chLines chan<- int, chAlarm chan<- parser.Alarm) {
	allWorkers = make(map[string]bool)
	workersCanWaitOnce := new(sync.Once)
	mutex := new(sync.Mutex) // mutex for all workers

	// main loop to watch for newly emerging data sources
	// when we start, they may not exist, but latter on, they come out suddenly
	for {
		for _, guard := range conf.Guards {
			if options.parser != "" && !guard.HasParser(options.parser) {
				// only one parser applied
				continue
			}

			for _, dataSource := range guardDataSources(guard) {
				if _, present := allWorkers[dataSource]; present {
					// this data source is already being guarded
					continue
				}

				wg.Add(1)
				allWorkers[dataSource] = true

				var newWoker NewWorker
				if guard.IsFileSource() {
					newWoker = newLogfileWorker
				} else if guard.IsDbSource() {
					newWoker = newDbWorker
				}

				var worker = newWoker(len(allWorkers),
					dataSource, guard, options.tailmode, wg, mutex, chLines, chAlarm)
				go worker.Run()
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
