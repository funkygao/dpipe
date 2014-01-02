package plugins

import (
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/pretty"
	"github.com/funkygao/tail"
	"os"
	"path/filepath"
	"time"
)

type LogfileInputConfig struct {
	DiscoverInterval int    `json:"discovery_interval"`
	Tail             bool   `json:"tail"`
	Glob             string `json:"glob"`
}

type LogfileInput struct {
	*LogfileInputConfig

	stopChan chan bool
}

func (this *LogfileInput) Init(config interface{}) {
	if globals.Debug {
		pretty.Printf("%# v\n", config)
	}

	conf := config.(*LogfileInputConfig)
	this.LogfileInputConfig = conf

	this.stopChan = make(chan bool)
}

func (this *LogfileInput) Config() interface{} {
	return LogfileInputConfig{
		DiscoverInterval: 5,
		Tail:             true,
	}

}

func (this *LogfileInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Logger.Printf("[%s] started\n", r.Name())
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
				globals.Logger.Printf("[%s] found new file input: %v\n", fn)
			}
			go this.runSingleLogfileInput(fn, r, e)
		}

		select {
		case <-reloadChan:
			// TODO

		case <-time.After(time.Duration(this.DiscoverInterval) * time.Second):

		case <-this.stopChan:
			if globals.Verbose {
				globals.Logger.Printf("[%s] stopped\n", r.Name())
			}
			stopped = true
		}
	}

	return nil
}

func (this *LogfileInput) runSingleLogfileInput(fn string, r engine.InputRunner, e *engine.EngineConfig) {
	var tailConf tail.Config
	if this.Tail {
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

	var pack *engine.PipelinePack
	inChan := r.InChan()
	globals := engine.Globals()
	for line := range t.Lines {
		if globals.Debug {
			globals.Logger.Printf("[%s] got line: %s\n", r.Name(), line)
		}

		pack = <-inChan
		pack.Message.FromLine(line.Text)
		r.Inject(pack)
	}
}

func (this *LogfileInput) Stop() {
	close(this.stopChan)
}

func (this *LogfileInput) inputs() []string {
	logfiles, err := filepath.Glob(this.Glob)
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
