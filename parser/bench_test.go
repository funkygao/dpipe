package parser

import (
	"testing"
)

func BenchmarkUUID(b *testing.B) {
	indexer := newIndexer(nil)
	m := make(map[string]int, 1000)
	for i := 0; i < b.N; i++ {
		uuid, err := indexer.genUUID()
		if err != nil {
			b.Fatalf("GenUUID error %s", err)
		}
		b.StopTimer()
		c := m[uuid]
		if c > 0 {
			b.Fatalf("duplicate uuid[%s] count %d", uuid, c)
		}

		m[uuid] = c + 1
		b.StartTimer()
	}
}
