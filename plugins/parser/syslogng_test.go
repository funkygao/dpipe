package parser

import (
	"github.com/funkygao/assert"
	"testing"
)

func TestParse(t *testing.T) {
	r, err := Parse("non-exist", "msg")
	assert.Equal(t, ErrInvaidParser, err)
	assert.Equal(t, "", r)
}

func TestSyslogStatsParser(t *testing.T) {
	msg := "Jan 26 21:30:37 ip-172-31-13-40 syslog-ng[2156]: Log statistics; dropped='program(/mnt/htdocs/als/forward.php)=0', dropped='program(/mnt/htdocs/als/bigdata_forward.php)=0', processed='center(queued)=1699039', processed='center(received)=1699039', processed='destination(d_als_prog)=1697379', processed='destination(d_boot)=0', processed='destination(d_auth)=0', processed='destination(d_cron)=1426', processed='destination(d_mlal)=0', processed='destination(d_als_file)=0', processed='destination(d_mesg)=229', processed='destination(d_bigdata_prog)=0', processed='destination(d_haproxy)=0', processed='destination(d_cons)=0', processed='destination(d_spol)=0', processed='destination(d_mail)=5', processed='source(s_udp)=0', processed='source(s_bigdata)=0', processed='source(s_sys)=1660', processed='source(s_als)=1697379', suppressed='program(/mnt/htdocs/als/forward.php)=0', suppressed='program(/mnt/htdocs/als/bigdata_forward.php)=0'"
	match, alarm, _ := parseSyslogNgStats(msg)
	assert.Equal(t, "", alarm)
	assert.Equal(t, true, match)

	msg = "Jan 26 21:30:37 ip-172-31-13-40 syslog-ng[2156]: Log statistics; dropped='program(/mnt/htdocs/als/forward.php)=56', dropped='program(/mnt/htdocs/als/bigdata_forward.php)=20', processed='center(queued)=1699039', processed='center(received)=1699039', processed='destination(d_als_prog)=1697379', processed='destination(d_boot)=0', processed='destination(d_auth)=0', processed='destination(d_cron)=1426', processed='destination(d_mlal)=0', processed='destination(d_als_file)=0', processed='destination(d_mesg)=229', processed='destination(d_bigdata_prog)=0', processed='destination(d_haproxy)=0', processed='destination(d_cons)=0', processed='destination(d_spol)=0', processed='destination(d_mail)=5', processed='source(s_udp)=0', processed='source(s_bigdata)=0', processed='source(s_sys)=1660', processed='source(s_als)=1697379', suppressed='program(/mnt/htdocs/als/forward.php)=0', suppressed='program(/mnt/htdocs/als/bigdata_forward.php)=0'"
	match, alarm, _ = parseSyslogNgStats(msg)
	assert.Equal(t, true, match)
	assert.Equal(t, " [/mnt/htdocs/als/forward.php]dropped:56 [/mnt/htdocs/als/bigdata_forward.php]dropped:20", alarm)
}
