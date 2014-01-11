package plugins

import (
	"fmt"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	sky "github.com/funkygao/skyapi"
)

type SkyOutput struct {
	client   *sky.Client
	stopChan chan bool
}

func (this *SkyOutput) Init(config *conf.Conf) {
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

func (this *SkyOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
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

	if globals.Verbose {
		globals.Printf("%s stopped\n", r.Name())
	}

	return nil
}

func init() {
	engine.RegisterPlugin("SkyOutput", func() engine.Plugin {
		return new(SkyOutput)
	})
}
