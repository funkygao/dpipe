package rule

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestLoadRuleEngine(t *testing.T) {
	c, err := LoadRuleEngine("fixture/alser.cf")
	t.Log(err)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, c)

	assert.Equal(t, 2, len(c.Workers))
	assert.Equal(t, 1, len(c.Parsers))

	assert.Equal(t, true, c.IsParserApplied("MemcacheFailParser"))
	assert.Equal(t, false, c.IsParserApplied("NonExistParser"))

	// guards
	assert.Equal(t, "/mnt/funplus/logs/fp_rstory/memcache_to.*.log", c.Workers[0].TailGlob)
	assert.Equal(t, "/mnt/funplus/logs/fp_rstory/history/cache_set_fail*", c.Workers[1].HistoryGlob)
	assert.Equal(t, false, c.Workers[0].HasParser("NonExistParser"))
	assert.Equal(t, true, c.Workers[0].HasParser("MemcacheFailParser"))

	// parsers
	p := c.Parsers["MemcacheFailParser"]
	assert.Equal(t, "Line", p.Class)
	assert.Equal(t, "MemcacheFailParser", p.Id)
	assert.Equal(t, []string{"FgYellow"}, p.Colors)

	// parser keys
	assert.Equal(t, 2, len(p.Fields))
	assert.Equal(t, "key", p.Fields[0].Name)
	assert.Equal(t, "string", p.Fields[0].Type)
	assert.Equal(t, "timeout", p.Fields[1].Name)
	assert.Equal(t, "int", p.Fields[1].Type)
	assert.Equal(t, "blah", p.Fields[1].Contain)
	assert.Equal(t, []string{"digit", "token"}, p.Fields[0].Regex)

	// get parser by id
	mp := c.ParserById("MemcacheFailParser")
	assert.Equal(t, "Line", mp.Class)

	np := c.ParserById("NonExistsParser")
	if np != nil {
		t.Error("expected nil, got ", np)
	}

}
