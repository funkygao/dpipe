package plugins

import (
	"fmt"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/observer"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/tail"
	"os"
	"path/filepath"
	"strings"
)

type logfileSource struct {
	glob    string
	excepts []string
	project string
	sink    int

	_files []string
}

func (this *logfileSource) load(config *conf.Conf) {
	this.glob = config.String("glob", "")
	if this.glob == "" {
		panic("glob cannot be empty")
	}

	this.project = config.String("proj", "")
	this.excepts = config.StringList("except", nil)
	this.sink = config.Int("sink", 0)
	this._files = make([]string, 0, 200)
}

func (this *logfileSource) refresh() {
	files, err := filepath.Glob(this.glob)
	if err != nil {
		panic(err)
	}

	this._files = this._files[:0]
	for _, fn := range files {
		basename := filepath.Base(fn)
		for _, except := range this.excepts {
			if strings.HasPrefix(basename, except) {
				continue
			}
		}

		this._files = append(this._files, fn)
	}
}

type AlsLogInput struct {
	stopChan chan bool
	sources  []*logfileSource
}

func (this *AlsLogInput) Init(config *conf.Conf) {
	if engine.Globals().Debug {
		engine.Globals().Printf("%#v\n", *config)
	}

	this.stopChan = make(chan bool)

	// get the sources
	this.sources = make([]*logfileSource, 0, 200)
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

func (this *AlsLogInput) CleanupForRestart() {

}

func (this *AlsLogInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	var (
		reloadChan  = make(chan interface{})
		openedFiles = make(map[string]bool)
		stopped     = false
	)

	observer.Subscribe(engine.RELOAD, reloadChan)

	for !stopped {
		this.refreshSources()

		for _, source := range this.sources {
			for _, fn := range source._files {
				if _, present := openedFiles[fn]; present {
					continue
				}

				if globals.Verbose {
					globals.Printf("found new file %s\n", fn)
				}

				openedFiles[fn] = true
				go this.runSingleAlsLogInput(fn, r, e, *source, &stopped)
			}
		}

		select {
		case <-reloadChan:
			// TODO

		case <-r.Ticker():

		case <-this.stopChan:
			if globals.Verbose {
				globals.Printf("%s stopped\n", r.Name())
			}
			stopped = true
		}
	}

	return nil
}

func (this *AlsLogInput) runSingleAlsLogInput(fn string, r engine.InputRunner,
	e *engine.EngineConfig, source logfileSource, stopped *bool) {
	var tailConf tail.Config
	if engine.Globals().Tail {
		tailConf = tail.Config{
			Follow:   true, // tail -f
			ReOpen:   true, // tail -F
			Poll:     true, // Poll for file changes instead of using inotify
			Location: &tail.SeekInfo{Offset: int64(0), Whence: os.SEEK_END},
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
		globals = engine.Globals()
	)

	for line := range t.Lines {
		if globals.Debug {
			globals.Printf("[%s]got line: %s\n", fn, line.Text)
		}

		if *stopped {
			break
		}

		pack = <-inChan
		if err := pack.Message.FromLine(line.Text); err != nil {
			e.Project(source.project).Printf("[%s]%v <= %s\n", fn, err, line.Text)
			pack.Recycle()
			continue
		}

		pack.Project = source.project
		pack.Logfile.SetPath(fn)
		pack.Message.Sink = source.sink
		r.Inject(pack)
	}
}

func (this *AlsLogInput) refreshSources() {
	for _, s := range this.sources {
		s.refresh()
	}
}

func init() {
	engine.RegisterPlugin("AlsLogInput", func() engine.Plugin {
		return new(AlsLogInput)
	})
}
