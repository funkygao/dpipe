package plugins

import (
	"github.com/funkygao/tail"
	"regexp"
	"strings"
	"testing"
)

func BenchmarkStringConstains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strings.Contains("we are all here", "all")
	}
}

func BenchmarkStringConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "home" + ":" + "nice"
	}
}

func BenchmarkStringLen(b *testing.B) {
	s := "abcdefg"
	for i := 0; i < b.N; i++ {
		_ = len(s)
	}
}

func BenchmarkRegexpMatch(b *testing.B) {
	pattern := "child \\d+ started"
	line := "adfasdf  asdfas dfasdf child 12 started with asdfasf"
	for i := 0; i < b.N; i++ {
		regexp.MatchString(pattern, line)
	}
}

func BenchmarkRegexpMatchCompiled(b *testing.B) {
	pattern := regexp.MustCompile("child \\d+ started")
	line := "adfasdf  asdfas dfasdf child 12 started with asdfasf"
	for i := 0; i < b.N; i++ {
		pattern.MatchString(line)
	}
}

func BenchmarkTailFile(b *testing.B) {
	var tailConf tail.Config
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		t, err := tail.TailFile("/var/log/system.log", tailConf)
		if err != nil {
			panic(err)
		}
		b.StartTimer()
		<-t.Lines
	}

}
