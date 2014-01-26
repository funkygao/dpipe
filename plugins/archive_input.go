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
	stopping    bool
	runner      engine.InputRunner
	h           engine.PluginHelper
	chkpnt      *als.FileCheckpoint
	workersWg   *sync.WaitGroup
	lineN       int64
	workerNChan chan int
	rootDir     string
	ignores     []string
	ident       string
	project     string
	leftN       int32
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
	this.ignores = config.StringList("ignores", nil)
}

func (this *ArchiveInput) CleanupForRestart() bool {
	this.chkpnt.Dump()
	return false
}

func (this *ArchiveInput) Stop() {
	this.stopping = true
}

func (this *ArchiveInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	this.runner = r
	this.h = h

	this.chkpnt.Load()
	go func() {
		for !this.stopping {
			select {
			case <-r.Ticker():
				this.chkpnt.Dump()

				if this.leftN > 0 {
					engine.Globals().Printf("[%s]Left %d files", r.Name(), this.leftN)
				}
			}
		}
	}()

	this.workersWg = new(sync.WaitGroup)

	filepath.Walk(this.rootDir, this.setLeftN)
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("Total files:%d", this.leftN)
	}

	// do the real job
	filepath.Walk(this.rootDir, this.runSingleLogfile)

	// wait for all workers done
	this.workersWg.Wait()
	this.chkpnt.Dump()

	if globals.Verbose {
		globals.Printf("[%s]Total msg: %d", r.Name(), this.lineN)
	}

	return nil
}

func (this *ArchiveInput) shouldRunSingleLogfile(path string) bool {
	if this.chkpnt.Contains(path) {
		return false
	}

	for _, ignore := range this.ignores {
		if strings.HasPrefix(filepath.Base(path), ignore) {
			return false
		}
	}

	if strings.HasSuffix(filepath.Base(path), ".gz") {
		return false
	}

	return true
}

func (this *ArchiveInput) setLeftN(path string, f os.FileInfo,
	err error) (e error) {
	if f == nil || f.IsDir() || !this.shouldRunSingleLogfile(path) {
		return
	}

	this.leftN += 1
	return nil
}

func (this *ArchiveInput) runSingleLogfile(path string, f os.FileInfo,
	err error) (e error) {
	if f == nil || f.IsDir() || !this.shouldRunSingleLogfile(path) {
		return
	}
	if (f.Mode() & os.ModeSymlink) > 0 {
		// are we to handle this case?
		// TODO maybe it will never happens
	}

	this.workersWg.Add(1)

	// limit concurrent workers
	this.workerNChan <- 1

	go this.doRunSingleLogfile(path)

	return
}

func (this *ArchiveInput) doRunSingleLogfile(path string) {
	reader := als.NewAlsReader(path)
	if e := reader.Open(); e != nil {
		panic(e)
	}

	defer func() {
		reader.Close()
		this.workersWg.Done()
		atomic.AddInt32(&this.leftN, -1)

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

	for !this.stopping {
		line, err = reader.ReadLine()
		switch err {
		case nil:
			lineN += 1
			atomic.AddInt64(&this.lineN, 1)
			if globals.Verbose && lineN == 1 {
				project.Printf("[%s]started\n", path)
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
