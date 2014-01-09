package plugins

import (
	"github.com/funkygao/assert"
	"testing"
)

func TestNormalize(t *testing.T) {
	msg := "batch token error! pre: leuw53.1e2t2j; current: 2m2w1z.1e2sz5 (1)"
	b := normalizers["batch_token"].ReplaceAll([]byte(msg), []byte("?"))
	assert.Equal(t, "batch token error! ?", string(b))

	msg = "user id: 34343434 adfasf"
	m := normalizers["digit"].ReplaceAll([]byte(msg), []byte("?"))
	assert.Equal(t, "user id: ? adfasf", string(m))
}
