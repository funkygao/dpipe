package engine

import (
	"fmt"
	"github.com/funkygao/als"
	"sync"
	"sync/atomic"
	"time"
)

// Main pipeline data structure containing a AlsMessage and other metadata
type PipelinePack struct {
	// Where to put back myself when reference count zeros
	RecycleChan chan *PipelinePack

	// Reference counter, internal GC
	RefCount int32

	// Decoded msg
	Message *als.AlsMessage

	// AlsMessage only knows a line, we need data source for routing
	Logfile *als.AlsLogfile

	Input bool

	// For routing
	Ident string

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
		Input:        false,
		diagnostics:  NewPacketTracking(),
		Message:      als.NewAlsMessage(),
		Logfile:      als.NewAlsLogfile(),
	}
}

func (this *PipelinePack) IncRef() {
	atomic.AddInt32(&this.RefCount, 1)
}

func (this PipelinePack) String() string {
	s := fmt.Sprintf("%s@%s, rc=%d, loop=%d, runner=%v", this.Ident, this.Project,
		this.RefCount, this.MsgLoopCount)
	if this.Logfile != nil {
		s = fmt.Sprintf("%s src=%s", s, this.Logfile.Base())
	}
	if this.Message != nil {
		s = fmt.Sprintf("%s payload=%s", s, this.Message.Payload)
	}
	if this.EsIndex != "" {
		s = fmt.Sprintf("%s, index{%s, %s}", s, this.EsIndex, this.EsType)
	}
	if this.CardinalityKey != "" {
		s = fmt.Sprintf("%s, cardinal{%s, %s, %v}", s, this.CardinalityKey,
			this.CardinalityInterval, this.CardinalityData)
	}
	if this.diagnostics != nil {
		s = fmt.Sprintf("%s, lastAccess=%s", s, time.Since(this.diagnostics.LastAccess))
	}

	return s
}

func (this *PipelinePack) Reset() {
	this.RefCount = int32(1)
	this.MsgLoopCount = 0
	this.Project = ""
	this.EsIndex = ""
	this.EsType = ""
	this.Input = false
	this.CardinalityKey = ""
	this.CardinalityData = nil
	this.CardinalityInterval = ""
	this.Ident = ""
	this.diagnostics.Reset()
	this.Message.Reset()
}

func (this *PipelinePack) CopyTo(that *PipelinePack) {
	that.Project = this.Project
	that.Ident = this.Ident
	that.Input = this.Input
	that.EsIndex = this.EsIndex
	that.EsType = this.EsType
	that.CardinalityKey = this.CardinalityKey
	that.CardinalityData = this.CardinalityData
	that.CardinalityInterval = this.CardinalityInterval

	that.Message = this.Message.QuickClone()
	that.Logfile.SetPath(this.Logfile.Path())
}

/*
                    IncRef2
                   -------> Output -------------> DecRef2 --+
                  |                                         |
         rc=1     | IncRef3                                 |
Input -> packA -> |-------> Filter1 -> Output1 -> DecRef1 --| race
                  |                                         | condition
                  | IncRef4                                 |
                  |-------> Filter2 -> Output2 -> DecRef0 --+
                  |                                  |
                  | DecRef3                          V
                   -------- router                   o[recyled]
*/
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
	mutex      sync.Mutex

	// Records the plugins the packet has been handed to
	pluginRunners []PluginRunner
}

func NewPacketTracking() *PacketTracking {
	return &PacketTracking{LastAccess: time.Now(),
		mutex:         sync.Mutex{},
		pluginRunners: make([]PluginRunner, 0, 8)}
}

func (this *PacketTracking) AddStamp(pluginRunner PluginRunner) {
	this.mutex.Lock()
	this.pluginRunners = append(this.pluginRunners, pluginRunner)
	this.mutex.Unlock()
	this.LastAccess = time.Now()
}

func (this *PacketTracking) Reset() {
	this.mutex.Lock()
	this.pluginRunners = this.pluginRunners[:0] // a tip in golang to avoid re-alloc
	this.mutex.Unlock()
	this.LastAccess = time.Now()
}

func (this *PacketTracking) Runners() []PluginRunner {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.pluginRunners
}

func (this *PacketTracking) RunnerCount() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return len(this.pluginRunners)

}
