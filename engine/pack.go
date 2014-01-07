package engine

import (
	"github.com/funkygao/als"
	"sync/atomic"
	"time"
)

// Main pipeline data structure containing a AlsMessage and other metadata
type PipelinePack struct {
	// Raw data yet to be decoded
	MsgBytes []byte

	// Decoded msg
	Message *als.AlsMessage

	Logfile *als.AlsLogfile

	// used for ES _type
	Typ string

	// Where to put back myself when reference count zeros
	RecycleChan chan *PipelinePack
	RefCount    int32

	// Project name
	Project string

	// To avoid infinite message loops
	MsgLoopCount int

	// Used internally to stamp diagnostic information
	diagnostics *PacketTracking
}

func NewPipelinePack(recycleChan chan *PipelinePack) (this *PipelinePack) {
	return &PipelinePack{
		RecycleChan:  recycleChan,
		RefCount:     int32(1),
		MsgLoopCount: 0,
		diagnostics:  NewPacketTracking(),
		Message:      als.NewAlsMessage(),
		Logfile:      als.NewAlsLogfile(),
	}
}

func (this *PipelinePack) IncRef() {
	atomic.AddInt32(&this.RefCount, 1)
}

func (this *PipelinePack) Reset() {
	this.RefCount = int32(1)
	this.MsgLoopCount = 0
	this.Project = ""
	this.Typ = ""
	this.diagnostics.Reset()
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

type PacketTracking struct {
	LastAccess time.Time

	// Records the plugins the packet has been handed to
	pluginRunners []PluginRunner
}

func NewPacketTracking() *PacketTracking {
	return &PacketTracking{time.Now(), make([]PluginRunner, 0, 8)}
}

func (this *PacketTracking) AddStamp(pluginRunner PluginRunner) {
	this.pluginRunners = append(this.pluginRunners, pluginRunner)
	this.LastAccess = time.Now()
}

func (this *PacketTracking) Reset() {
	this.pluginRunners = this.pluginRunners[:0] // a tip in golang to avoid re-alloc
	this.LastAccess = time.Now()
}

func (this *PacketTracking) PluginNames() (names []string) {
	names = make([]string, 0, 4)
	for _, pr := range this.pluginRunners {
		names = append(names, pr.Name())
	}

	return
}

func (this *PacketTracking) Runners() []PluginRunner {
	return this.pluginRunners
}
