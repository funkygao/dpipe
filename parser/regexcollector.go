package parser

import (
	"github.com/funkygao/alser/config"
	"regexp"
)

// Raw msg can be regex'ed
type RegexCollectorParser struct {
	CollectorParser
	r *regexp.Regexp
}

func newRegexCollectorParser(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) (this *RegexCollectorParser) {
	this = new(RegexCollectorParser)
	this.init(conf, chUpstream, chDownstream)
	if this.conf.MsgRegex == "" {
		panic(this.id() + ": empty msg_regex")
	}
	this.r = regexp.MustCompile(this.conf.MsgRegex)

	go this.CollectAlarms()

	return
}

func (this *RegexCollectorParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.CollectorParser.ParseLine(line)

	matches := this.r.FindAllStringSubmatch(msg, 10000)
	if len(matches) == 0 {
		if debug {
			logger.Printf("%s invalid msg: %s\n", this.id(), msg)
		}

		return
	}

	vals := matches[0]
	if len(vals) <= this.conf.MsgRegexKeys[len(this.conf.MsgRegexKeys)-1] {
		if debug {
			logger.Printf("%s invalid msg: %s\n", this.id(), msg)
		}

		return
	}

	args := make([]interface{}, 0)
	for _, idx := range this.conf.MsgRegexKeys {
		args = append(args, vals[idx])
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{area, ts}, args...)
	this.insert(args...)

	return
}
