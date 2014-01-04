package plugins

import (
	"fmt"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/observer"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/tail"
	"os"
	"path/filepath"
)

type logfileSource struct {
	glob    string
	files   []string
	project string
	nexts   []string
}

func (this *logfileSource) validate() {
	if this.glob == "" {
		panic("AlsLogInput.sources.glob cannot be empty")
	}
}

type AlsLogInput struct {
	stopChan chan bool
	sources  []logfileSource
}

func (this *AlsLogInput) Init(config *conf.Conf) {
	globals := engine.Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *config)
	}

	this.stopChan = make(chan bool)

	// get the sources
	this.sources = make([]logfileSource, 0, 200)
	for i := 0; i < len(config.List("sources", nil)); i++ {
		keyPrefix := fmt.Sprintf("sources[%d].", i)
		source := logfileSource{}
		source.glob = config.String(keyPrefix+"glob", "")
		source.project = config.String(keyPrefix+"proj", "")
		source.nexts = config.StringList(keyPrefix+"nexts", nil)
		source.validate()
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
			for _, fn := range source.files {
				if _, present := openedFiles[fn]; present {
					continue
				}

				openedFiles[fn] = true
				if globals.Debug {
					globals.Printf("[%s] found new file input: %v\n", fn)
				}

				go this.runSingleAlsLogInput(fn, r, e, &stopped, source.project, source.nexts)
			}
		}

		select {
		case <-reloadChan:
			// TODO

		case <-r.Ticker():

		case <-this.stopChan:
			if globals.Verbose {
				globals.Printf("[%s] stopped\n", r.Name())
			}
			stopped = true
		}
	}

	return nil
}

func (this *AlsLogInput) runSingleAlsLogInput(fn string, r engine.InputRunner,
	e *engine.EngineConfig, stopped *bool, project string, nexts []string) {
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
			globals.Printf("[%s]got line: %s\n", r.Name(), line.Text)
		}

		if *stopped {
			break
		}

		pack = <-inChan
		if err := pack.Message.FromLine(line.Text); err != nil {
			e.Project(project).Printf("%v <= %s\n", err, line.Text)
			continue
		}

		pack.Project = project
		pack.Logfile.SetPath(fn)
		pack.Nexts = nexts
		r.Inject(pack)
	}
}

func (this *AlsLogInput) refreshSources() {
	var err error
	for idx, source := range this.sources {
		this.sources[idx].files, err = filepath.Glob(source.glob)
		if err != nil {
			panic(err)
		}
	}
}

func init() {
	engine.RegisterPlugin("AlsLogInput", func() engine.Plugin {
		return new(AlsLogInput)
	})
}
