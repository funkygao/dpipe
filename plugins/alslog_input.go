package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/golib/sortedmap"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/tail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type logfileSource struct {
	glob     string
	ignores  []string
	project  string
	ident    string
	disabled bool
	tail     bool

	_files []string
}

func (this *logfileSource) load(config *conf.Conf) {
	this.glob = config.String("glob", "")
	if this.glob == "" {
		panic("glob cannot be empty")
	}

	this.project = config.String("project", "")
	this.tail = config.Bool("tail", true)
	this.ignores = config.StringList("ignores", nil)
	this.disabled = config.Bool("disabled", false)
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
	this._files = make([]string, 0, 100)
}

func (this *logfileSource) ignored(fn string) bool {
	for _, ignore := range this.ignores {
		if strings.HasPrefix(filepath.Base(fn), ignore) {
			return true
		}
	}

	return false
}

func (this *logfileSource) refresh(wg *sync.WaitGroup) {
	defer wg.Done()

	if this.disabled {
		return
	}

	files, err := filepath.Glob(this.glob)
	if err != nil {
		panic(err)
	}

	this._files = this._files[:0]
	for _, fn := range files {
		if !this.ignored(fn) {
			this._files = append(this._files, fn)
		}
	}
}

type AlsLogInput struct {
	stopChan     chan bool
	showProgress bool
	counters     *sortedmap.SortedMap // ident -> N
	sources      []*logfileSource
}

func (this *AlsLogInput) Init(config *conf.Conf) {
	if engine.Globals().Debug {
		engine.Globals().Printf("%#v\n", *config)
	}

	this.showProgress = config.Bool("show_pregress", true)
	this.counters = sortedmap.NewSortedMap()
	this.stopChan = make(chan bool)

	// get the sources
	this.sources = make([]*logfileSource, 0, 20)
	for i := 0; i < len(config.List("sources", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("sources[%d]", i))
		if err != nil {
			panic(err)
		}

		source := new(logfileSource)
		source.load(section)
		this.sources = append(this.sources, source)
	}
}

func (this *AlsLogInput) Stop() {
	close(this.stopChan)
}

func (this *AlsLogInput) CleanupForRestart() bool {
	return true
}

func (this *AlsLogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	var (
		reloadChan = make(chan interface{})
		stopped    = false
		opened     = make(map[string]bool) // safe because within a goroutine
	)

	observer.Subscribe(engine.RELOAD, reloadChan)

	for !stopped {
		this.refreshSources()

		for _, source := range this.sources {
			for _, fn := range source._files {
				if _, present := opened[fn]; present {
					continue
				}

				opened[fn] = true
				go this.runSingleAlsLogInput(fn, r, h, *source, &stopped)
			}
		}

		select {
		case <-reloadChan:
			// TODO

		case <-r.Ticker():
			this.handlePeriodicalCounters(len(opened))

		case <-this.stopChan:
			stopped = true
		}
	}

	// FIXME wait for all

	return nil
}

func (this *AlsLogInput) handlePeriodicalCounters(opendFiles int) {
	if !this.showProgress {
		return
	}

	var (
		n       = 0
		globals = engine.Globals()
	)
	globals.Printf("Opened files: %d", opendFiles)
	for _, ident := range this.counters.SortedKeys() {
		if n = this.counters.Get(ident); n > 0 {
			globals.Printf("%-15s %8d", ident, n)

			this.counters.Set(ident, 0)
		}
	}
}

func (this *AlsLogInput) runSingleAlsLogInput(fn string, r engine.InputRunner,
	h engine.PluginHelper, source logfileSource, stopped *bool) {
	var tailConf tail.Config
	if source.tail {
		tailConf = tail.Config{
			LimitRate: int64(0), // lines per second
			Follow:    true,     // tail -f
			ReOpen:    true,     // tail -F
			Poll:      true,     // Poll for file changes instead of using inotify
			Location:  &tail.SeekInfo{Offset: int64(0), Whence: os.SEEK_END},
		}
	}

	t, err := tail.TailFile(fn, tailConf)
	if err != nil {
		panic(err)
	}
	defer t.Stop()

	var (
		pack      *engine.PipelinePack
		inChan    = r.InChan()
		line      *tail.Line
		ok        bool
		checkStop = time.Duration(time.Second)
		globals   = engine.Globals()
	)

	if globals.Verbose {
		globals.Printf("[%s]%s started", source.project, fn)
	}

LOOP:
	for !*stopped {
		select {
		case line, ok = <-t.Lines:
			if !ok {
				break LOOP
			}

			if globals.Debug {
				globals.Printf("[%s]got line: %s\n", filepath.Base(fn), line.Text)
			}

			pack = <-inChan
			if err := pack.Message.FromLine(line.Text); err != nil {
				project := h.Project(source.project)
				if project.ShowError && err != als.ErrEmptyLine {
					project.Printf("[%s]%v: %s", fn, err, line.Text)
				}

				pack.Recycle()
				continue
			}

			this.counters.Inc(source.ident, 1)
			pack.Project = source.project
			pack.Logfile.SetPath(fn)
			pack.Ident = source.ident
			r.Inject(pack)

		case <-time.After(checkStop):

		}
	}

	if globals.Verbose {
		globals.Printf("[%s]%s stopped", source.project, fn)
	}

}

func (this *AlsLogInput) refreshSources() {
	wg := new(sync.WaitGroup)
	for _, s := range this.sources {
		if s.disabled {
			continue
		}

		wg.Add(1)
		go s.refresh(wg)
	}

	wg.Wait()
}

func init() {
	engine.RegisterPlugin("AlsLogInput", func() engine.Plugin {
		return new(AlsLogInput)
	})
}
