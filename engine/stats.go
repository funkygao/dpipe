package engine

import (
	"runtime"
)

type EngineStats struct {
	engine   *EngineConfig
	MemStats *runtime.MemStats
}

func newEngineStats(e *EngineConfig) (this *EngineStats) {
	this = new(EngineStats)
	this.engine = e
	this.MemStats = new(runtime.MemStats)
	return
}

func (this *EngineStats) NumGoroutine() int {
	return runtime.NumGoroutine()
}

func (this *EngineStats) LastGC() uint64 {
	this.refreshMemStats()
	return this.MemStats.LastGC
}

func (this *EngineStats) MemInfo() runtime.MemStats {
	this.refreshMemStats()
	return *this.MemStats
}

func (this *EngineStats) refreshMemStats() {
	runtime.ReadMemStats(this.MemStats)
}
