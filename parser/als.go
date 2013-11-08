/*

           AlsParser
               |
        ------------------------------------
       |               |                    |
   JsonLineParser  CollectorParser   HostLineParser
                       |
                   ----------------
                  |                |
         JsonCollectorParser   PhperrorCollectorParser

*/
package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
	"os"
	"strconv"
	"strings"
)

type logInfo struct {
	uid          int64
	snsid        string
	level        int
	payment_cash int
	uri          string
	scriptId     int64
	serial       int
	host         string // aws instance ip
	ip           string // remote user ip
}

// Parent parser for all
type AlsParser struct {
	Parser

	conf              *config.ConfParser
	chUpstreamAlarm   chan<- Alarm  // TODO not used yet
	chDownstreamAlarm chan<- string // consumed by parser itself

	color string
}

func (this *AlsParser) init(conf *config.ConfParser, chUpstream chan<- Alarm, chDownstream chan<- string) {
	this.conf = conf
	this.chUpstreamAlarm = chUpstream
	this.chDownstreamAlarm = chDownstream

	// setup color
	this.color = ""
	for _, c := range this.conf.Colors {
		this.color += COLOR_MAP[c]
	}
}

// Each ALS log line is area,timestamp,msg
// Most msg are json struct while some are raw text
func (this *AlsParser) ParseLine(line string) (area string, ts uint64, msg string) {
	if !this.conf.Enabled {
		return
	}

	fields := strings.SplitN(line, LINE_SPLITTER, LINE_SPLIT_NUM)
	area = fields[0]
	var err error
	ts, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}
	ts /= 1000 // raw timestamp is in ms

	msg = fields[2]
	return
}

func (this *AlsParser) Stop() {
}

func (this *AlsParser) Wait() {
}

func (this *AlsParser) id() string {
	return this.conf.Id
}

func (this *AlsParser) keysCount() int {
	return len(this.conf.Keys)
}

func (this *AlsParser) msgToJson(msg string) (data *json.Json) {
	var err error
	data, err = json.NewJson([]byte(msg))
	checkError(this.id(), err)

	return
}

func (this *AlsParser) jsonValue(data *json.Json, key, typ string) (val interface{}, err error) {
	switch typ {
	case "string":
		val, err = data.Get(key).String()
	case "int":
		val, err = data.Get(key).Int()
	case "float":
		val, err = data.Get(key).Float64()
	}

	return
}

// Extract values of json according config keys
func (this *AlsParser) valuesOfKeys(data *json.Json) (values []interface{}) {
	var err error
	var val interface{}
	values = make([]interface{}, 0)

	for _, key := range this.conf.Keys {
		keyParts := strings.SplitN(key.Name, ".", 2) // only 1 dot permitted
		if len(keyParts) > 1 {
			subData := data.Get(keyParts[0])
			val, err = this.jsonValue(subData, keyParts[1], key.Type)
		} else {
			val, err = this.jsonValue(data, key.Name, key.Type)
		}
		if err != nil {
			return
		}

		if key.Contain != "" && !strings.Contains(val.(string), key.Contain) {
			return
		}

		if !key.Visible {
			continue
		}

		if key.Ignores != nil {
			for _, ignore := range key.Ignores {
				if strings.Contains(val.(string), ignore) {
					return
				}
			}
		}

		if key.Regex != nil {
			logger.Println(key.Regex)
			for _, regex := range key.Regex {
				switch regex {
				case "digit":
					logger.Println(val)
					val = this.normalizeDigit(val.(string))
					logger.Println(val)
				case "token":
					val = this.normalizeBatchToken(val.(string))
				}
			}
		}

		values = append(values, val)
	}

	return
}

func (this *AlsParser) normalizeDigit(msg string) string {
	r := digitsRegexp.ReplaceAll([]byte(msg), []byte("?"))
	return string(r)
}

func (this *AlsParser) normalizeBatchToken(msg string) string {
	r := batchTokenRegexp.ReplaceAll([]byte(msg), []byte("pre cur"))
	return string(r)
}

func (this *AlsParser) colorPrintfLn(format string, args ...interface{}) {
	if daemonize {
		return
	}

	msg := fmt.Sprintf(format, args...)
	fmt.Println(this.color + msg + COLOR_MAP["Reset"])
}

func (this *AlsParser) alarmf(format string, args ...interface{}) {
	this.chDownstreamAlarm <- fmt.Sprintf(format, args...)
}

func (this *AlsParser) beep() {
	if daemonize {
		return
	}

	fmt.Print("\a")
	if beeped > MAX_BEEP_VISUAL_HINT {
		beeped = MAX_BEEP_VISUAL_HINT
	}
	fmt.Fprintln(os.Stderr, strings.Repeat("☹ ", beeped))
	beeped += 1
}
