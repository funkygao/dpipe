package main

import (
	"github.com/funkygao/alser/rule"
	"sync"
)

type SysWorker struct {
	Worker
	Lines chan string
}

func newSysWorker(id int,
	dataSource string, conf rule.ConfWorker, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int) Runnable {
	this := new(SysWorker)
	this.Worker = Worker{id: id,
		dataSource: dataSource, conf: conf, tailMode: tailMode,
		wg: wg, Mutex: mutex,
		chLines: chLines}

	return this
}

func (this *SysWorker) Run() {
	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}

}
