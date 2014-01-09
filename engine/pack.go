package engine

import (
	"github.com/funkygao/als"
	"sync/atomic"
	"time"
)

// Main pipeline data structure containing a AlsMessage and other metadata
type PipelinePack struct {
	// Where to put back myself when reference count zeros
	RecycleChan chan *PipelinePack

	// Reference counter, internal GC
	RefCount int32

	// Raw data yet to be decoded
	// TODO raw blob mainly for TCP stream
	MsgBytes []byte

	// Decoded msg
	Message *als.AlsMessage

	// AlsMessage only knows a line, we need data source for routing
	Logfile *als.AlsLogfile

	// For routing
	Sink string
	Tag  string

	// Project name
	Project string

	// To avoid infinite message loops
	MsgLoopCount int

	// Used internally to stamp diagnostic information
	diagnostics *PacketTracking

	EsType  string
	EsIndex string

	CardinalityKey      string
	CardinalityData     interface{}
	CardinalityInterval string
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
	this.EsIndex = ""
	this.EsType = ""
	this.CardinalityKey = ""
	this.CardinalityData = nil
	this.CardinalityInterval = ""
	this.Sink = ""
	this.Tag = ""
	this.diagnostics.Reset()
	this.Message.Reset()
}

func (this *PipelinePack) Recycle() {
	count := atomic.AddInt32(&this.RefCount, -1)
	if count == 0 {
		this.Reset()

		// reuse this pack to avoid re-alloc
		this.RecycleChan <- this
	} else if count < 0 {
		Globals().Panic("reference count below zero")
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
