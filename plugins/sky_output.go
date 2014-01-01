package plugins

import (
	"errors"
	"fmt"
	"github.com/funkygao/funpipe/engine"
	sky "github.com/funkygao/skyapi"
)

type SkyOutputConfig struct {
	Host string
	Port int
}

type SkyOutput struct {
	*SkyOutputConfig
	*sky.Client

	stopChan chan bool
}

func (this *SkyOutput) Init(config interface{}) {
	conf := config.(SkyOutputConfig)
	this.SkyOutputConfig = conf
	this.stopChan = make(chan bool)

	this.Client = sky.NewClient(this.Host)
	this.Client.Port = this.Port
	if !this.Ping() {
		panic(fmt.Sprintf("sky server not running: %s:%d", this.Host, this.Port))
	}
}

func (this *SkyOutput) Config() interface{} {
	return SkyOutputConfig{
		Host: "localhost",
		Port: 8585,
	}
}

func (this *SkyOutput) Run(r engine.OutputRunner, c *engine.EngineConfig) error {
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
	engine.RegisterPlugin("SkyOutput", func() interface{} {
		return new(SkyOutput)
	})
}
