package parser

import (
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/rule"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Raw msg can be regex'ed
type RegexCollectorParser struct {
	CollectorParser
	r *regexp.Regexp
}

func newRegexCollectorParser(conf *config.ConfParser, chDownstream chan<- string) (this *RegexCollectorParser) {
	this = new(RegexCollectorParser)
	this.init(conf, chDownstream)
	if this.conf.MsgRegex == "" {
		panic(this.id() + ": empty msg_regex")
	}
	this.r = regexp.MustCompile(this.conf.MsgRegex)

	go this.CollectAlarms()

	return
}

func (this *RegexCollectorParser) ParseLine(line string) (area string, ts uint64, msg string) {
	area, ts, msg = this.CollectorParser.ParseLine(line)
	if msg == "" {
		if verbose {
			logger.Printf("got empty msg: %s\n", line)
		}

		return
	}

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
			if err != nil {
				if debug {
					logger.Printf("%s invalid msg: %s\n", this.id(), msg)
				}

				return
			}
		}

		args = append(args, val)
	}

	if this.conf.Indexing {
		indexJson, _ := json.NewJson([]byte("{}"))
		indexJson.Set(INDEX_COL_AREA, area)
		indexJson.Set(INDEX_COL_TIMESTAMP, ts)
		indexJson.Set("msg", msg)

		date := time.Unix(int64(ts), 0)
		indexer.c <- indexEntry{indexName: this.conf.IndexName, typ: this.conf.Title, date: &date, data: indexJson}
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{area, ts}, args...)
	this.insert(args...)

	return
}
