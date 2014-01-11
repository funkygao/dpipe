package plugins

import (
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type BatchAlsLogInput struct {
	runner      engine.InputRunner
	project     string
	h           engine.PluginHelper
	chkpnt      *als.FileCheckpoint
	workerNChan chan int
	rootDir     string
	workersWg   *sync.WaitGroup
	excepts     []string
	sink        string
}

func (this *BatchAlsLogInput) Init(config *conf.Conf) {
	this.rootDir = config.String("root_dir", "")
	this.sink = config.String("sink", "")
	if this.sink == "" {
		panic("empty sink")
	}
	this.project = config.String("proj", "rs")
	this.workerNChan = make(chan int, config.Int("concurrent_num", 20))
	this.chkpnt = als.NewFileCheckpoint(config.String("chkpntfile", ""))
	this.excepts = config.StringList("except", nil)
}

func (this *BatchAlsLogInput) CleanupForRestart() {
	this.chkpnt.Dump()
}

func (this *BatchAlsLogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	this.runner = r
	this.h = h

	this.chkpnt.Load()
	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for _ = range ticker.C {
			this.chkpnt.Dump()
		}
	}()

	this.workersWg = new(sync.WaitGroup)

	filepath.Walk(this.rootDir, this.runSingleLogfile)

	// wait for all workers done
	this.workersWg.Wait()
	ticker.Stop()
	this.chkpnt.Dump()

	globals.Printf("%s done, whole system is shutdown...", r.Name())
	globals.Shutdown()

	return nil
}

func (this *BatchAlsLogInput) Stop() {

}

func (this *BatchAlsLogInput) shouldRunSingleLogfile(path string) bool {
	if this.chkpnt.Contains(path) {
		return false
	}

	for _, ex := range this.excepts {
		if strings.HasPrefix(filepath.Base(path), ex) {
			return false
		}
	}

	return true
}

func (this *BatchAlsLogInput) runSingleLogfile(path string, f os.FileInfo, err error) (e error) {
	if f == nil || f.IsDir() || !this.shouldRunSingleLogfile(path) {
		return
	}

	// limit concurrent workers
	this.workerNChan <- 1
	go this.doRunSingleLogfile(path)

	return
}

func (this *BatchAlsLogInput) doRunSingleLogfile(path string) {
	this.workersWg.Add(1)

	reader := als.NewAlsReader(path)
	if e := reader.Open(); e != nil {
		panic(e)
	}

	defer func() {
		reader.Close()
		this.workersWg.Done()

		<-this.workerNChan // release the lock
	}()

	var (
		line    []byte
		lineN   int
		inChan  = this.runner.InChan()
		err     error
		project = this.h.Project(this.project)
		pack    *engine.PipelinePack
		globals = engine.Globals()
	)

LOOP:
	for !globals.Stopping {
		line, err = reader.ReadLine()
		switch err {
		case nil:
			lineN += 1
			if globals.Verbose && lineN == 1 {
				project.Printf("[%s]started\n", path)
			}
			if globals.Debug {
				project.Printf("[%s]#%d\n", path, lineN)
			}

			pack = <-inChan
			if err = pack.Message.FromLine(string(line)); err != nil {
				project.Printf("[%s]error: %v <= %s\n", path, err, string(line))

				pack.Recycle()
				continue
			}

			pack.Sink = this.sink
			pack.Project = this.project
			pack.Logfile.SetPath(path)
			this.runner.Inject(pack)

		case io.EOF:
			if globals.Verbose {
				project.Printf("[%s]done, lines: %d\n", path, lineN)
			}

			this.chkpnt.Put(path)
			this.chkpnt.Dump()

			break LOOP
		}
	}

}

func init() {
	engine.RegisterPlugin("BatchAlsLogInput", func() engine.Plugin {
		return new(BatchAlsLogInput)
	})
}
