package plugins

import (
	"github.com/funkygao/dpipe/engine"
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

func (this *NetReceiverInput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	listener, err := net.Listen("tcp4", this.listenAddr)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	go func() {
		globals := engine.Globals()
		for _ = range r.Tick() {
			globals.Printf("Total %dB, speed: %s/s", this.totalBytes, this.periodBytes)
		}
	}()

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
	buf := make([]byte, 4<<10)
	inChan := r.InChan()

	var (
		pack    *engine.PipelinePack
		globals = engine.Globals()
		ok      bool
	)

	globals.Printf("Connection from %s", conn.RemoteAddr().String())

	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			conn.Close()
			break
		}

		atomic.AddInt64(this.totalBytes, n)

		pack, ok = <-inChan
		if !ok {
			break
		}

		pack.Bytes = buf[:n] // copy bytes
		r.Inject(pack)

	}

	globals.Printf("Closed connection from %s", conn.RemoteAddr().String())

}

func init() {
	engine.RegisterPlugin("NetReceiverInput", func() engine.Plugin {
		return new(NetReceiverInput)
	})
}
