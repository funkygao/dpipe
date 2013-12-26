package worker

import (
	"bitbucket.org/bertimus9/systemstat"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/alser/rule"
	"sync"
	"time"
)

func init() {
	RegisterWorkerPlugin("sys", func() interface{} {
		return new(SysWorker)
	})
}

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

func (s *stats) json() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	dst := new(bytes.Buffer)
	dst.Write(b)
	return dst.String(), nil
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
	defer this.Done()

	stats := newStats()
	var line string
	for {
		stats.GatherStats(true)

		jsonStats, err := stats.json()
		if err != nil {
			if options.verbose {
				logger.Printf("%s got invalid sysstat: %v\n", *this, *stats)
			}

			continue
		}

		if options.debug {
			logger.Printf("%s got line: %s\n", *this, line)
		}

		for _, parserId := range this.conf.Parsers {
			line = fmt.Sprintf("als,%d,%s",
				time.Now().Unix(), jsonStats)
			parser.Dispatch(parserId, line)
		}

		time.Sleep(10 * time.Second)
	}

	if options.verbose {
		logger.Printf("%s finished\n", *this)
	}

}
