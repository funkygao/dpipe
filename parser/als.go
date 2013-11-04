/*

           AlsParser
               |
        ---------------
       |               |
   JsonLineParser  CollectorParser
                       |
                   ----------------
                  |                |
         JsonCollectorParser    RawLineCollectorParser

*/
package parser

import (
	"errors"
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
	chUpstreamAlarm   chan<- Alarm // TODO not used yet
	chDownstreamAlarm chan<- string

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

func (this *AlsParser) ParseLine(line string) (area string, ts uint64, msg string) {
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

func (this *AlsParser) parseJsonLine(line string) (area string, ts uint64, data *json.Json) {
	var (
		msg string
		err error
	)
	area, ts, msg = this.ParseLine(line)
	data, err = json.NewJson([]byte(msg))
	checkError(err)

	return
}

func (this *AlsParser) extractDataValue(data *json.Json, name, typ string) (val interface{}, err error) {
	switch typ {
	case "string":
		val, err = data.Get(name).String()
	case "int":
		val, err = data.Get(name).Int()
	case "float":
		val, err = data.Get(name).Float64()
	}

	return
}

func (this *AlsParser) extractDataValues(data *json.Json) (values []interface{}, err error) {
	values = make([]interface{}, 0)
	for _, key := range this.conf.Keys {
		var val interface{}

		keyParts := strings.SplitN(key.Name, ".", 2) // only 1 dot permitted
		if len(keyParts) > 1 {
			subData := data.Get(keyParts[0])
			val, err = this.extractDataValue(subData, keyParts[1], key.Type)
		} else {
			val, err = this.extractDataValue(data, key.Name, key.Type)
		}

		if err != nil {
			return
		}

		if key.MustBe != "" && key.MustBe != val.(string) {
			err = errors.New("must be:" + key.MustBe + ", got:" + val.(string))
			return
		}

		if key.Ignores != nil {
			for _, ignore := range key.Ignores {
				if strings.Contains(val.(string), ignore) {
					err = errors.New("ignored:" + val.(string))
					return
				}
			}
		}

		if key.Regex != nil {
			for key.Regex !=nil; _, regex := range key.Regex {
				switch regex {
				case "digit":
					val = this.normalizeDigit(val)
				case "token":
					val = this.normalizeBatchToken(val)
				}
			}
		}

		values = append(values, val)
	}

	return
}

func (this *AlsParser) Stop() {
}

func (this *AlsParser) Wait() {
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
