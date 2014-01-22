package plugins

import (
	"bufio"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/gofmt"
	conf "github.com/funkygao/jsconf"
	"net"
	"sync/atomic"
)

// Receive log content line by line via TCP
type NetReceiverInput struct {
	listenAddr  string
	maxLineSize int
	totalBytes  int64
	periodBytes int64
}

func (this *NetReceiverInput) Init(config *conf.Conf) {
	this.listenAddr = config.String("listen_addr", ":9000")
	this.maxLineSize = config.Int("max_line_size", 8<<10)
}

func (this *NetReceiverInput) reportStats(r engine.OutputRunner) {
	globals := engine.Globals()

	for _ = range r.Tick() {
		globals.Printf("Total %s, speed: %s/s",
			gofmt.ByteSize(this.totalBytes),
			gofmt.ByteSize(this.periodBytes))

		this.periodBytes = int64(0)
	}
}

func (this *NetReceiverInput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	listener, err := net.Listen("tcp4", this.listenAddr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	go this.reportStats(r)

LOOP:
	for {
		conn, err := listener.Accept()
		if err != nil {
			engine.Globals().Println(err)
			break LOOP
		}

		go this.handleTcpConnection(conn, r)
	}
}

func (this *NetReceiverInput) handleTcpConnection(conn net.Conn, r engine.OutputRunner) {
	var (
		lineReader = bufio.NewReader(conf)
		line       string
		err        error
		pack       *engine.PipelinePack
		ok         bool
		inChan     = r.InChan()
		globals    = engine.Globals()
	)

	globals.Printf("Connection from %s", conn.RemoteAddr())

	for {
		line, err = lineReader.ReadString('\n')
		if err != nil {
			globals.Printf("[%s]%s", conn.RemoteAddr(), err)
			continue
		}

		atomic.AddInt64(this.totalBytes, len(line))
		atomic.AddInt64(this.periodBytes, len(line))

		pack, ok = <-inChan
		if !ok {
			break
		}

		// TODO marshal the pack from line
		r.Inject(pack)
	}

	globals.Printf("Closed connection from %s", conn.RemoteAddr().String())

}

func init() {
	engine.RegisterPlugin("NetReceiverInput", func() engine.Plugin {
		return new(NetReceiverInput)
	})
}
