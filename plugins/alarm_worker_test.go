package plugins

import (
	"github.com/funkygao/als"
	"github.com/funkygao/assert"
	conf "github.com/funkygao/jsconf"
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

func TestAlarmFieldIgnore(t *testing.T) {
	f := alarmWorkerConfigField{}
	c, _ := conf.Load("fixture/ignore.cf")
	f.init(c)
	assert.Equal(t, 3, len(f._regexIgnores))
	assert.Equal(t, 3, len(f.ignores))
	assert.Equal(t, "msg", f.name)
	assert.Equal(t, "syslogngStats", f.parser)
	assert.Equal(t, als.KEY_TYPE_STRING, f.typ)
	assert.Equal(t, 0, len(f.normalizers))

	msg := als.NewAlsMessage()
	line := `ae,1391857296,{"host":"172.31.32.91","msg":"2014-02-08 06:01:35 [INFO] Update local version to e9818f0812b3933ac62d630b5b99aac5 for release royal.ae.php"}`
	msg.FromLine(line)
	assert.Equal(t, "ae", msg.Area)
	value, err := f.value(msg)
	assert.Equal(t, nil, err)
	assert.Equal(t, "2014-02-08 06:01:35 [INFO] Update local version to e9818f0812b3933ac62d630b5b99aac5 for release royal.ae.php", value)
}
