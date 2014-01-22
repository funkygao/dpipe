package engine

import (
	"sync/atomic"
	"testing"
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
