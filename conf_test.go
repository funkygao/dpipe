package main

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	defer cleanup()

	conf := loadJsonConfig("fixtures/alser.test.json")
	assert.Equal(t, 6, len(conf))
	assert.Equal(t, "payments", conf[5].Name)
	assert.Equal(t, "/mnt/funplus/logs/fp_rstory/cache_set_fail.*.log", conf[0].Pattern)
}

func TestJsonConfigParsers(t *testing.T) {
	defer cleanup()

	conf := loadJsonConfig("fixtures/alser.test.json")
	assert.Equal(t, 3, len(conf.parsers()))
}
