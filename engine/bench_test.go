package engine

import (
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"
)

func BenchmarkAtomicAdd64(b *testing.B) {
	var x int64
	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&x, 5)
	}
}

func BenchmarkAtomicAdd32(b *testing.B) {
	var x int32
	for i := 0; i < b.N; i++ {
		atomic.AddInt32(&x, 5)
	}
}

func BenchmarkRouterStatsUpdate(b *testing.B) {
	pack := NewPipelinePack(nil)
	stat := routerStats{}
	pack.Message.FromLine(`us,1389913256544,{"uid":9837688,"date":20140117,"ip":"90.165.137.106"}`)
	for i := 0; i < b.N; i++ {
		stat.update(pack)
	}
}

func BenchmarkMatcher(b *testing.B) {
	pack := NewPipelinePack(nil)
	pack.Ident = "foox"
	matcher := NewMatcher([]string{"foo", "bar", "ping", "pong"}, nil)
	for i := 0; i < b.N; i++ {
		matcher.match(pack)
	}
}

func BenchmarkMakeChan(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make(chan interface{})
	}
}

func BenchmarkGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		go func() {
			runtime.Goexit()
		}()
	}
}

func BenchmarkRecycleChannel(b *testing.B) {
	b.ReportAllocs()

	var (
		globals          = DefaultGlobals()
		inputRecycleChan = make(chan *PipelinePack, globals.RecyclePoolSize)
		pack             *PipelinePack
	)

	for i := 0; i < globals.RecyclePoolSize; i++ {
		inputPack := NewPipelinePack(inputRecycleChan)
		inputRecycleChan <- inputPack
	}

	for i := 0; i < b.N; i++ {
		pack = <-inputRecycleChan
		pack.Recycle()
	}

	b.SetBytes(int64(unsafe.Sizeof(pack)))
}

func BenchmarkPluginChannel(b *testing.B) {
	recycleChan := make(chan *PipelinePack, 150)
	pack := NewPipelinePack(recycleChan)
	go func(pack *PipelinePack) {
		for {
			recycleChan <- pack
		}
	}(pack)
	for i := 0; i < b.N; i++ {
		<-recycleChan
	}
	b.SetBytes(int64(unsafe.Sizeof(pack)))
}

func BenchmarkPackCopyTo(b *testing.B) {
	b.ReportAllocs()

	line := `us,1389913256544,{"uid":9837688,"date":20140117,"ip":"90.165.137.106"}`
	pack := NewPipelinePack(nil)
	pack.Message.FromLine(line)
	p := NewPipelinePack(nil)
	for i := 0; i < b.N; i++ {
		pack.CopyTo(p)
	}

	b.SetBytes(int64(len(line)))
}
