package engine

import (
	"github.com/funkygao/als"
	"sync/atomic"
)

// Main Heka pipeline data structure containing raw message data, a Message
// object, and other Heka related message metadata.
type PipelinePack struct {
	// Used for storage of binary blob data that has yet to be decoded into a
	// Message object.
	MsgBytes []byte

	Message *als.AlsMessage

	RecycleChan chan *PipelinePack

	Decoded bool

	RefCount int32

	MsgLoopCount uint
	// Used internally to stamp diagnostic information onto a packet
	diagnostics *PacketTracking
}

func NewPipelinePack(recycleChan chan *PipelinePack) (pack *PipelinePack) {
	pack = &PipelinePack{
		RefCount:     uint32(1),
		MsgLoopCount: uint(0),
		Decoded:      false,
		RecycleChan:  recycleChan,
	}

}

func (this *PipelinePack) Reset() {
	this.RefCount = 1
	this.MsgLoopCount = 0
}

func (this *PipelinePack) Recycle() {
	count := atomic.AddInt32(&this.RefCount, -1)
	if count == 0 {
		this.Reset()
		this.RecycleChan <- this
	}
}
