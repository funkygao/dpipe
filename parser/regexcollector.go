package parser

import (
	"github.com/funkygao/alser/config"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	args := make([]interface{}, 0)
	for _, key := range this.conf.MsgRegexKeys {
		parts := strings.SplitN(key, ":", 2) // idx:typ
		idx, _ := strconv.Atoi(parts[0])
		var val interface{} = vals[idx]
		if len(parts) > 1 {
			var err error
			switch parts[1] {
			case "float":
				val, err = strconv.ParseFloat(val.(string), 64)
			case "int", "money":
				val, err = strconv.Atoi(val.(string))
			case "base_file":
				val = filepath.Base(val.(string))
			}
		}

		args = append(args, val)
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{area, ts}, args...)
	this.insert(args...)

	return
}
