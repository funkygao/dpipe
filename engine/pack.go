package engine

import (
	"github.com/funkygao/als"
	"sync/atomic"
)

// Main pipeline data structure containing a AlsMessage and other metadata
type PipelinePack struct {
	// Raw data yet be decoded to AlsMessage obj TODO
	MsgBytes []byte

	// AlsMessage obj pointer
	Message *als.AlsMessage

	RecycleChan chan *PipelinePack
	RefCount    int32

	// Routing table
	Nexts []string
	// Project name
	Project string

	// To avoid infinite message loops
	MsgLoopCount int
}

func NewPipelinePack(recycleChan chan *PipelinePack) (this *PipelinePack) {
	return &PipelinePack{
		RecycleChan:  recycleChan,
		RefCount:     int32(1),
		MsgLoopCount: 0,
		Message:      als.NewAlsMessage(),
	}
}

func (this *PipelinePack) Reset() {
	this.RefCount = int32(1)
	this.MsgLoopCount = 0

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
