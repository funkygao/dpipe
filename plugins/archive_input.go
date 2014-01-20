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
	"sync/atomic"
)

type ArchiveInput struct {
	runner      engine.InputRunner
	h           engine.PluginHelper
	chkpnt      *als.FileCheckpoint
	workersWg   *sync.WaitGroup
	lineN       int64
	workerNChan chan int
	rootDir     string
	excepts     []string
	ident       string
	project     string
	leftN       int
}

func (this *ArchiveInput) Init(config *conf.Conf) {
	this.rootDir = config.String("root_dir", "")
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
	this.project = config.String("project", "rs")
	this.workerNChan = make(chan int, config.Int("concurrent_num", 20))
	this.chkpnt = als.NewFileCheckpoint(config.String("chkpntfile", ""))
	this.excepts = config.StringList("except", nil)
}

func (this *ArchiveInput) CleanupForRestart() bool {
	this.chkpnt.Dump()
	return false
}

func (this *ArchiveInput) Stop() {

}

func (this *ArchiveInput) showProgress(r engine.InputRunner) {
	for {
		select {
		case <-r.Ticker():
			if this.leftN == 0 {
				return
			}

			engine.Globals().Printf("[%s]Left %d files", r.Name(), this.leftN)
		}
	}

}

func (this *ArchiveInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	this.runner = r
	this.h = h

	this.chkpnt.Load()
	go func() {
		for _ = range r.Ticker() {
			this.chkpnt.Dump()
		}
	}()

	this.workersWg = new(sync.WaitGroup)

	filepath.Walk(this.rootDir, this.runSingleLogfile)

	go this.showProgress(r)

	// wait for all workers done
	this.workersWg.Wait()
	this.chkpnt.Dump()

	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s]Total msg: %d", r.Name(), this.lineN)
	}

	return nil
}

func (this *ArchiveInput) shouldRunSingleLogfile(path string) bool {
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

func (this *ArchiveInput) runSingleLogfile(path string, f os.FileInfo, err error) (e error) {
	if f == nil || f.IsDir() || !this.shouldRunSingleLogfile(path) {
		return
	}

	this.workersWg.Add(1)
	this.leftN += 1

	go this.doRunSingleLogfile(path)

	return
}

func (this *ArchiveInput) doRunSingleLogfile(path string) {
	// limit concurrent workers
	this.workerNChan <- 1

	reader := als.NewAlsReader(path)
	if e := reader.Open(); e != nil {
		panic(e)
	}

	defer func() {
		reader.Close()
		this.workersWg.Done()
		this.leftN -= 1

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

	for {
		line, err = reader.ReadLine()
		switch err {
		case nil:
			lineN += 1
			atomic.AddInt64(&this.lineN, 1)
			if globals.Verbose && lineN == 1 {
				project.Printf("[%s]started\n", path)
			}
			if globals.Debug {
				project.Printf("[%s]#%d\n", path, lineN)
			}

			pack = <-inChan
			if err = pack.Message.FromLine(string(line)); err != nil {
				if project.ShowError && err != als.ErrEmptyLine {
					project.Printf("[%s]%v: %s", path, err, string(line))
				}

				pack.Recycle()
				continue
			}

			pack.Ident = this.ident
			pack.Project = this.project
			pack.Logfile.SetPath(path)
			if globals.Debug {
				globals.Println(*pack)
			}
			this.runner.Inject(pack)

		case io.EOF:
			if globals.Verbose {
				project.Printf("[%s]done, lines: %d\n", path, lineN)
			}

			this.chkpnt.Put(path)
			this.chkpnt.Dump()

			return

		default:
			// unknown error
			panic(err)
		}
	}

}

func init() {
	engine.RegisterPlugin("ArchiveInput", func() engine.Plugin {
		return new(ArchiveInput)
	})
}
