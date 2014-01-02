package plugins

import (
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/observer"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/tail"
	"os"
	"path/filepath"
	"time"
)

type LogfileInput struct {
	discoverInterval time.Duration
	stopChan         chan bool
}

func (this *LogfileInput) Init(config *conf.Conf) {
	globals := engine.Globals()
	if globals.Debug {
		globals.Printf("%#v\n", *config)
	}

	this.discoverInterval = time.Duration(config.Int("discovery_interval", 5))
	this.stopChan = make(chan bool)
}

func (this *LogfileInput) Stop() {
	close(this.stopChan)
}

func (this *LogfileInput) CleanupForRestart() {

}

func (this *LogfileInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {
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
		for _, fn := range this.inputs() {
			if _, present := openedFiles[fn]; present {
				continue
			}

			openedFiles[fn] = true
			if globals.Debug {
				globals.Printf("[%s] found new file input: %v\n", fn)
			}
			go this.runSingleLogfileInput(fn, r, e, &stopped)
		}

		select {
		case <-reloadChan:
			// TODO

		case <-time.After(this.discoverInterval * time.Second):

		case <-this.stopChan:
			if globals.Verbose {
				globals.Printf("[%s] stopped\n", r.Name())
			}
			stopped = true
		}
	}

	return nil
}

func (this *LogfileInput) runSingleLogfileInput(fn string, r engine.InputRunner,
	e *engine.EngineConfig, stopped *bool) {
	var tailConf tail.Config
	if true {
		tailConf = tail.Config{
			Follow:   true, // tail -f
			ReOpen:   true, // tail -F
			Poll:     true, // Poll for file changes instead of using inotify
			Location: &tail.SeekInfo{Offset: int64(0), Whence: os.SEEK_END},
			//MustExist: false,
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
		pack.Message.FromLine(line.Text)
		r.Inject(pack)
	}
}

func (this *LogfileInput) inputs() []string {
	logfiles, err := filepath.Glob("this.Glob")
	if err != nil {
		panic(err)
	}

	return logfiles
}

func init() {
	engine.RegisterPlugin("LogfileInput", func() engine.Plugin {
		return new(LogfileInput)
	})
}
