package main

import (
	"fmt"
	"github.com/funkygao/alser/rule"
	"sync"
	"time"
)

type NewWorker func(id int, dataSource string,
	conf rule.ConfWorker, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int) Runnable

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
	conf       rule.ConfWorker
	tailMode   bool

	*sync.Mutex
	wg      *sync.WaitGroup
	chLines chan<- int
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

func invokeWorkers(ruleEngine *rule.RuleEngine, wg *sync.WaitGroup, workersCanWait chan<- bool,
	chLines chan<- int) {
	allWorkers = make(map[string]bool)
	workersCanWaitOnce := new(sync.Once)
	mutex := new(sync.Mutex) // mutex for all workers

	// main loop to watch for newly emerging data sources
	// when we start, they may not exist, but latter on, they come out suddenly
	for {
		for _, w := range ruleEngine.Workers {
			if options.parser != "" && !w.HasParser(options.parser) {
				// only one parser applied
				continue
			}
			if !w.Enabled {
				continue
			}

			for _, dataSource := range guardDataSources(w) {
				if _, present := allWorkers[dataSource]; present {
					// this data source is already being guarded
					continue
				}

				if options.debug {
					logger.Printf("data source added: %s", dataSource)
				}

				wg.Add(1)
				allWorkers[dataSource] = true

				var newWoker NewWorker
				if w.Type == rule.DATASOURCE_FILE {
					newWoker = newLogfileWorker
				} else if w.Type == rule.DATASOURCE_DB {
					newWoker = newDbWorker
				} else if w.Type == rule.DATASOURCE_SYS {
					newWoker = newSysWorker
				}

				var worker = newWoker(len(allWorkers),
					dataSource, w, options.tailmode, wg, mutex, chLines)
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
