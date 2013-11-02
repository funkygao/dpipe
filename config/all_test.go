package config

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	c, err := LoadConfig("fixture/alser.cf")
	t.Log(err)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, c)

	assert.Equal(t, 2, len(c.guards))
	assert.Equal(t, 1, len(c.parsers))

	assert.Equal(t, "/mnt/funplus/logs/fp_rstory/memcache_to.*.log", c.guards[0].tailLogGlob)
	assert.Equal(t, "/mnt/funplus/logs/fp_rstory/history/cache_set_fail*", c.guards[1].historyLogGlob)

	assert.Equal(t, "Line", c.parsers[0].class)
	assert.Equal(t, "MemcacheFailParser", c.parsers[0].id)
	assert.Equal(t, []string{"key", "timeout"}, c.parsers[0].lineColumns)
	assert.Equal(t, []string{"FgYellow"}, c.parsers[0].colors)
}
