package main

import (
	"bitbucket.org/bertimus9/systemstat"
	"bytes"
	"fmt"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/alser/rule"
	"sync"
	"time"
)

type SysWorker struct {
	Worker
	Lines chan string
}

type stats struct {
	startTime time.Time

	// stats this process
	ProcUptime        float64 //seconds
	ProcMemUsedPct    float64
	ProcCPUAvg        systemstat.ProcCPUAverage
	LastProcCPUSample systemstat.ProcCPUSample `json:"-"`
	CurProcCPUSample  systemstat.ProcCPUSample `json:"-"`

	// stats for whole system
	LastCPUSample systemstat.CPUSample `json:"-"`
	CurCPUSample  systemstat.CPUSample `json:"-"`
	SysCPUAvg     systemstat.CPUAverage
	SysMemK       systemstat.MemSample
	LoadAverage   systemstat.LoadAvgSample
	SysUptime     systemstat.UptimeSample

	// bookkeeping
	procCPUSampled bool
	sysCPUSampled  bool
}

func newStats() *stats {
	s := stats{}
	s.startTime = time.Now()
	return &s
}

func (s *stats) GatherStats(percent bool) {
	s.SysUptime = systemstat.GetUptime()
	s.ProcUptime = time.Since(s.startTime).Seconds()

	s.SysMemK = systemstat.GetMemSample()
	s.LoadAverage = systemstat.GetLoadAvgSample()

	s.LastCPUSample = s.CurCPUSample
	s.CurCPUSample = systemstat.GetCPUSample()

	if s.sysCPUSampled { // we need 2 samples to get an average
		s.SysCPUAvg = systemstat.GetCPUAverage(s.LastCPUSample, s.CurCPUSample)
	}
	// we have at least one sample, subsequent rounds will give us an average
	s.sysCPUSampled = true

	s.ProcMemUsedPct = 100 * float64(s.CurProcCPUSample.ProcMemUsedK) / float64(s.SysMemK.MemTotal)

	s.LastProcCPUSample = s.CurProcCPUSample
	s.CurProcCPUSample = systemstat.GetProcCPUSample()
	if s.procCPUSampled {
		s.ProcCPUAvg = systemstat.GetProcCPUAverage(s.LastProcCPUSample, s.CurProcCPUSample, s.ProcUptime)
	}
	s.procCPUSampled = true
}

func (s *stats) json() string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	dst := new(bytes.Buffer)
	dst.Write(b)
	return dst.String()
}

func newSysWorker(id int,
	dataSource string, conf config.ConfGuard, tailMode bool,
	wg *sync.WaitGroup, mutex *sync.Mutex,
	chLines chan<- int, chAlarm chan<- parser.Alarm) Runnable {
	this := new(SysWorker)
	this.Worker = Worker{id: id,
		dataSource: dataSource, conf: conf, tailMode: tailMode,
		wg: wg, Mutex: mutex,
		chLines: chLines, chAlarm: chAlarm}
	this.Lines = make(chan string)

	return this
}

func (this *DbWorker) Run() {
	defer this.Done()

	stats := newStats()
	var line string
	for {
		stats.GatherStats(true)

		for _, parserId := range this.conf.Parsers {
			line = fmt.Sprintf("sys,%d,%s",
				time.Now().Unix(), stats.json())
			parser.Dispatch(parserId, line)
		}

		time.Sleep(10 * time.Second)
	}

	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}

}
