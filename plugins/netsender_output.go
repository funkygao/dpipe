package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"net"
)

type netSenderTarget struct {
	c net.Conn
}

func (this *netSenderTarget) send(pack *engine.PipelinePack) {
	n, err := this.c.Write(pack.Bytes)
	if n != len(pack.Bytes) || err != nil {
		panic(err)
	}
}

// Send local logs contents to remote TCP line by line
type NetSenderOutput struct {
	remoteAddr  string
	maxLineSize int
	targets     map[string]*netSenderTarget
}

func (this *NetSenderOutput) Init(config *conf.Conf) {
	this.remoteAddr = config.String("remote_addr", ":9000")
	this.maxLineSize = config.Int("max_line_size", 8<<10)
}

func (this *NetSenderOutput) target(pack *engine.PipelinePack) *netSenderTarget {
	key := pack.Project + ":" + pack.Logfile.Base()
	if t, present := this.targets[key]; present {
		return t
	}

	// a new tcp conn
	remoteAddr, err := net.ResolveTCPAddr("tcp", this.remoteAddr)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		panic(err)
	}

	t := &netSenderTarget{c: conn}
	this.targets[key] = t
	return t
}

func (this *NetSenderOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

LOOP:
	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break LOOP
			}

			this.target(pack).send(pack)
			pack.Recycle()
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("NetSenderOutput", func() engine.Plugin {
		return new(NetSenderOutput)
	})
}
