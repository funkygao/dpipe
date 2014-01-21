package engine

import (
	"fmt"
	"github.com/funkygao/golib/gofmt"
	"runtime"
	"time"
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

func (this *EngineStats) Runtime() map[string]interface{} {
	s := make(map[string]interface{})
	this.refreshMemStats()
	s["goroutines"] = runtime.NumGoroutine()
	s["memory.allocated"] = gofmt.ByteSize(this.MemStats.Alloc)
	s["memory.mallocs"] = gofmt.ByteSize(this.MemStats.Mallocs)
	s["memory.frees"] = gofmt.ByteSize(this.MemStats.Frees)
	s["memory.gc.num"] = this.MemStats.NumGC
	s["memory.gc.total_pause"] = fmt.Sprintf("%dms",
		this.MemStats.PauseTotalNs/uint64(time.Millisecond))
	s["memory.heap"] = gofmt.ByteSize(this.MemStats.HeapAlloc)
	s["memory.heap.objects"] = gofmt.Comma(int64(this.MemStats.HeapObjects))
	s["memory.stack"] = gofmt.ByteSize(this.MemStats.StackInuse)
	gcPausesMs := make([]string, 0, 20)
	for _, pauseNs := range this.MemStats.PauseNs {
		if pauseNs == 0 {
			continue
		}

		gcPausesMs = append(gcPausesMs, fmt.Sprintf("%dms",
			pauseNs/uint64(time.Millisecond)))
	}
	s["memory.gc.pauses"] = gcPausesMs

	return s
}

func (this *EngineStats) refreshMemStats() {
	runtime.ReadMemStats(this.MemStats)
}
