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

	assert.Equal(t, 2, len(c.Guards))
	assert.Equal(t, 1, len(c.Parsers))

	// guards
	assert.Equal(t, "/mnt/funplus/logs/fp_rstory/memcache_to.*.log", c.Guards[0].TailLogGlob)
	assert.Equal(t, "/mnt/funplus/logs/fp_rstory/history/cache_set_fail*", c.Guards[1].HistoryLogGlob)

	// parsers
	p := c.Parsers[0]
	assert.Equal(t, "Line", p.Class)
	assert.Equal(t, "MemcacheFailParser", p.Id)
	assert.Equal(t, []string{"FgYellow"}, p.Colors)
	assert.Equal(t, "ALS Guard ", p.MailSubjectPrefix)
	assert.Equal(t, true, p.MailEnabled())
	assert.Equal(t, []string{"peng.gao@funplusgamenet.com", "zhengkai@gmail.com"}, p.MailRecipients)
	assert.Equal(t, "peng.gao@funplusgamenet.com,zhengkai@gmail.com", p.MailTos())
	// parser keys
	assert.Equal(t, 2, len(p.Keys))
	assert.Equal(t, "key", p.Keys[0].Key)
	assert.Equal(t, true, p.Keys[0].Required)
	assert.Equal(t, "timeout", p.Keys[1].Key)
	assert.Equal(t, false, p.Keys[1].Show)

}
