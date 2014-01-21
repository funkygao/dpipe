package plugins

import (
	"strings"
	"testing"
)

func BenchmarkStringConstains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strings.Contains("we are all here", "all")
	}
}
