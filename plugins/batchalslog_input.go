package plugins

import (
	"github.com/funkygao/als"
	"github.com/funkygao/funpipe/engine"
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
	chkpnt      *als.FileCheckpoint
	workerNChan chan int
	rootDir     string
	workersWg   *sync.WaitGroup
	excepts     []string
	sink        int
}

func (this *BatchAlsLogInput) Init(config *conf.Conf) {
	this.rootDir = config.String("root_dir", "")
	this.sink = config.Int("sink", 0)
	this.workerNChan = make(chan int, config.Int("concurrent_num", 20))
	this.chkpnt = als.NewFileCheckpoint(config.String("chkpntfile", ""))
	this.excepts = config.StringList("except", nil)
}

func (this *BatchAlsLogInput) CleanupForRestart() {
	this.chkpnt.Dump()
}

func (this *BatchAlsLogInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {
	this.runner = r

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

	engine.Globals().Printf("%s done", r.Name())

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
		this.chkpnt.Put(path)
		this.chkpnt.Dump()
		this.workersWg.Done()

		<-this.workerNChan // release the lock
	}()

	var (
		line    []byte
		lineN   int
		inChan  = this.runner.InChan()
		err     error
		pack    *engine.PipelinePack
		globals = engine.Globals()
	)

LOOP:
	for !globals.Stopping {
		line, err = reader.ReadLine()
		switch err {
		case nil:
			lineN += 1
			if globals.Verbose {
				globals.Printf("running %s, lineNum: %d\n", path, lineN)
			}

			pack = <-inChan
			if err = pack.Message.FromLine(string(line)); err != nil {
				globals.Printf("[%s]error line: %v <= %s\n", path, err, string(line))

				pack.Recycle()
				continue
			}

			pack.Message.Sink = this.sink
			this.runner.Inject(pack)

		case io.EOF:
			break LOOP
		}
	}

}

func init() {
	engine.RegisterPlugin("BatchAlsLogInput", func() engine.Plugin {
		return new(BatchAlsLogInput)
	})
}
