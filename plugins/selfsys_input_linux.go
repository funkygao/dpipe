package plugins

import (
	"bitbucket.org/bertimus9/systemstat"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"time"
)

// Analysis of current system stats
// Stats from /proc/uptime, /proc/loadavg, /proc/meminfo, /proc/stat
type SelfSysInput struct {
	stopChan chan bool
	sink     string
}

func (this *SelfSysInput) Init(config *conf.Conf) {
	this.stopChan = make(chan bool)
	this.sink = config.String("sink", "")
	if this.sink == "" {
		panic("empty sink")
	}
}

func (this *SelfSysInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	var (
		globals    = engine.Globals()
		stats      = newSysStat()
		inChan     = r.InChan()
		pack       *engine.PipelinePack
		jsonString string
		err        error
		stopped    = false
	)

	for !stopped {
		stats.gatherStats()
		jsonString, err = stats.jsonString()
		if err != nil {
			continue
		}

		pack = <-inChan
		if err = pack.Message.FromLine(fmt.Sprintf("als,%d,%s",
			time.Now().Unix(), jsonString)); err != nil {
			globals.Printf("invalid sys stat: %s\n", jsonString)
			pack.Recycle()
			continue
		}

		pack.Project = "als"
		pack.Sink = this.sink
		pack.EsIndex = "als"
		pack.EsType = "sys"
		r.Inject(pack)

		select {
		case <-this.stopChan:
			stopped = true

		case <-r.Ticker():
			// same effect as sleep
		}
	}

	return nil
}

func (this *SelfSysInput) Stop() {
	close(this.stopChan)
}

type sysStat struct {
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

func newSysStat() *sysStat {
	return &sysStat{startTime: time.Now()}
}

func (s *sysStat) gatherStats() {
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

func (s *sysStat) jsonString() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	dst := new(bytes.Buffer)
	dst.Write(b)
	return dst.String(), nil
}

func init() {
	engine.RegisterPlugin("SelfSysInput", func() engine.Plugin {
		return new(SelfSysInput)
	})
}
