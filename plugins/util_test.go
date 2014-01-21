package plugins

import (
	"github.com/funkygao/assert"
	"github.com/funkygao/dpipe/engine"
	"testing"
	"time"
)

func BenchmarkIndexName(b *testing.B) {
	date, _ := time.Parse("2006-01-02 15:04", "2011-01-19 22:15")
	p := &engine.ConfProject{IndexPrefix: "rs"}
	for i := 0; i < b.N; i++ {
		indexName(p, "@ymw", date)
	}
}

func TestIndexName(t *testing.T) {
	date, _ := time.Parse("2006-01-02 15:04", "2011-01-19 22:15")
	p := &engine.ConfProject{IndexPrefix: "rs"}
	assert.Equal(t, "fun_rs_2011_01", indexName(p, "@ym", date))
	assert.Equal(t, "fun_rs_2011_01_19", indexName(p, "@ymd", date))
	assert.Equal(t, "fun_rs_2011_w03", indexName(p, "@ymw", date))
	assert.Equal(t, "fun_foo", indexName(p, "foo", date))
}
