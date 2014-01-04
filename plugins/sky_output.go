package plugins

import (
	"fmt"
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
	sky "github.com/funkygao/skyapi"
)

type SkyOutput struct {
	client   *sky.Client
	stopChan chan bool
}

func (this *SkyOutput) Init(config *conf.Conf) {
	engine.Globals().Debugf("%#v\n", *config)

	this.stopChan = make(chan bool)
	var (
		host string = config.String("host", "localhost")
		port int    = config.Int("port", 8585)
	)
	this.client = sky.NewClient(host)
	this.client.Port = port

	if !this.client.Ping() {
		panic(fmt.Sprintf("sky server not running: %s:%d", host, port))
	}
}

func (this *SkyOutput) Run(r engine.OutputRunner, c *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	var (
		ok = true
	)

	for ok {
		select {
		case <-this.stopChan:
			ok = false

		default:
		}

	}

	return nil
}

func (this *SkyOutput) Stop() {
	close(this.stopChan)
}

func init() {
	engine.RegisterPlugin("SkyOutput", func() engine.Plugin {
		return new(SkyOutput)
	})
}
