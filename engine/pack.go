package engine

import (
	"github.com/funkygao/als"
	"sync/atomic"
)

type PipelinePack struct {
	// Raw data yet be decoded to AlsMessage obj TODO
	MsgBytes []byte

	// AlsMessage obj pointer
	Message *als.AlsMessage

	RecycleChan chan *PipelinePack

	// Recycle/reuse when zero
	RefCount int32

	// To avoid infinite message loops
	MsgLoopCount uint
}

func NewPipelinePack(recycleChan chan *PipelinePack) (this *PipelinePack) {
	this = &PipelinePack{
		RecycleChan:  recycleChan,
		RefCount:     int32(1),
		MsgLoopCount: uint(0),
		Message:      als.NewAlsMessage(),
	}

	return
}

func (this *PipelinePack) Reset() {
	this.RefCount = int32(1)
	this.MsgLoopCount = uint(0)

	this.Message.Reset()
}

func (this *PipelinePack) Recycle() {
	count := atomic.AddInt32(&this.RefCount, -1)
	if count == 0 {
		this.Reset()

		// reuse this pack to avoid re-alloc
		this.RecycleChan <- this
	}
}
