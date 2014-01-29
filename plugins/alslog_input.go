package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/golib/sortedmap"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/tail"
	"github.com/funkygao/tail/watch"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type logfileProject struct {
	name    string
	decode  bool
	sources []*logfileSource
}

func (this *logfileProject) load(config *conf.Conf) {
	this.name = config.String("name", "")
	if this.name == "" {
		panic("empty project name")
	}

	this.decode = config.Bool("decode", true)
	this.sources = make([]*logfileSource, 0, 10)
	for i := 0; i < len(config.List("sources", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("sources[%d]", i))
		if err != nil {
			panic(err)
		}

		source := new(logfileSource)
		source.project = this
		source.load(section)
		this.sources = append(this.sources, source)
	}
}

type logfileSource struct {
	glob  string // required
	ident string // required

	disabled bool
	tail     bool
	ignores  []string

	project *logfileProject

	_files []string // cache
}

func (this *logfileSource) load(config *conf.Conf) {
	this.glob = config.String("glob", "")
	if this.glob == "" {
		panic("glob cannot be empty")
	}
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
	this.tail = config.Bool("tail", true)
	this.ignores = config.StringList("ignores", nil)
	this.disabled = config.Bool("disabled", false)

	this._files = make([]string, 0, 50)
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
	projects     []*logfileProject
}

func (this *AlsLogInput) Init(config *conf.Conf) {
	if engine.Globals().Debug {
		engine.Globals().Printf("%#v\n", *config)
	}

	this.showProgress = config.Bool("show_progress", true)
	this.counters = sortedmap.NewSortedMap()
	this.stopChan = make(chan bool)
	watch.POLL_DURATION =
		time.Duration(config.Int("poll_interval_ms", 250)) * time.Millisecond

	this.projects = make([]*logfileProject, 0, 5)
	for i := 0; i < len(config.List("projects", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("projects[%d]", i))
		if err != nil {
			panic(err)
		}

		project := new(logfileProject)
		project.load(section)
		this.projects = append(this.projects, project)
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
		ever       = true
		opened     = make(map[string]bool) // safe because within a goroutine
	)

	observer.Subscribe(engine.RELOAD, reloadChan)

	for ever {
		this.refreshSources()

		for _, project := range this.projects {
			for _, source := range project.sources {
				for _, fn := range source._files {
					if _, present := opened[fn]; present {
						continue
					}

					opened[fn] = true
					go this.runSingleAlsLogInput(fn, r, h, *source)
				}
			}
		}

		select {
		case <-reloadChan:
			// TODO

		case <-r.Ticker():
			this.showPeriodicalStats(len(opened), r.TickLength())

		case <-this.stopChan:
			ever = false
		}
	}

	// FIXME wait for all

	return nil
}

func (this *AlsLogInput) showPeriodicalStats(opendFiles int, tl time.Duration) {
	if !this.showProgress {
		return
	}

	var (
		n       = 0
		total   = 0
		globals = engine.Globals()
	)
	globals.Printf("Opened files: %d, tickerLen: %s", opendFiles, tl)
	for _, ident := range this.counters.SortedKeys() {
		if n = this.counters.Get(ident); n > 0 {
			total += n
			globals.Printf("%-15s %12s messages", ident, gofmt.Comma(int64(n)))

			this.counters.Set(ident, 0)
		}
	}

	globals.Printf("%15s %12s", "Sum", gofmt.Comma(int64(total)))
}

func (this *AlsLogInput) runSingleAlsLogInput(fn string, r engine.InputRunner,
	h engine.PluginHelper, source logfileSource) {
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
		pack    *engine.PipelinePack
		inChan  = r.InChan()
		line    *tail.Line
		ok      bool
		globals = engine.Globals()
	)

	if this.showProgress {
		globals.Printf("[%s]%s started", source.project.name, fn)
	}

LOOP:
	for {
		select {
		case <-this.stopChan:
			break LOOP

		case line, ok = <-t.Lines:
			if !ok {
				break LOOP
			}

			this.counters.Inc(source.ident, 1)

			if globals.VeryVerbose {
				globals.Printf("[%s]got line: %s\n", filepath.Base(fn), line.Text)
			}

			pack = <-inChan
			pack.Project = source.project.name
			pack.Logfile.SetPath(fn)
			if source.project.decode {
				if err := pack.Message.FromLine(line.Text); err != nil {
					project := h.Project(source.project.name)
					if project.ShowError && err != als.ErrEmptyLine {
						project.Printf("[%s]%v: %s", fn, err, line.Text)
					}

					pack.Recycle()
					continue
				}
			} else {
				pack.Message.SetSize(len(line.Text))
			}

			pack.Ident = source.ident
			r.Inject(pack)
		}
	}

	if globals.Verbose {
		globals.Printf("[%s]%s stopped", source.project.name, fn)
	}
}

func (this *AlsLogInput) refreshSources() {
	wg := new(sync.WaitGroup)
	for _, project := range this.projects {
		for _, source := range project.sources {
			if source.disabled {
				continue
			}

			wg.Add(1)
			go source.refresh(wg)
		}
	}

	wg.Wait()
}

func init() {
	engine.RegisterPlugin("AlsLogInput", func() engine.Plugin {
		return new(AlsLogInput)
	})
}
