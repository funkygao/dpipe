package engine

import (
	"runtime"
)

type Stats struct {
	engine   *EngineConfig
	memStats *runtime.MemStats
}

func newStats(e *EngineConfig) (this *Stats) {
	this = new(Stats)
	this.engine = e
	this.memStats = new(runtime.MemStats)
	return
}

func (this *Stats) DispatchedMessageCount() int64 {
	return this.engine.router.totalProcessedMsgN
}

func (this *Stats) NumGoroutine() int {
	return runtime.NumGoroutine()
}

func (this *Stats) LastGC() uint64 {
	this.refreshMemStats()
	return this.memStats.LastGC
}

func (this *Stats) MemStats() runtime.MemStats {
	this.refreshMemStats()
	return *this.memStats
}

func (this *Stats) refreshMemStats() {
	runtime.ReadMemStats(this.memStats)
}
