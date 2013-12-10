/*

           AlsParser
               |
        ------------------------------------
       |               |                    |
   JsonLineParser  CollectorParser   HostLineParser
                       |
                   ----------------
                  |                |
         JsonCollectorParser

*/
package parser

import (
	"errors"
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/als"
	"github.com/funkygao/alser/config"
	"os"
	"path/filepath"
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

	var err error
	area, ts, msg, err = als.ParseAlsLine(line)
	if err != nil {
		if debug {
			logger.Printf("[%s]invalid line: %s", this.id(), line)
		}
	}

	return
}

func (this *AlsParser) Stop() {
}

func (this *AlsParser) Wait() {
}

func (this *AlsParser) id() string {
	return this.conf.Id
}

func (this *AlsParser) msgToJson(msg string) (data *json.Json, err error) {
	data, err = als.MsgToJson(msg)

	return
}

func (this *AlsParser) jsonValue(data *json.Json, key, typ string) (val interface{}, err error) {
	switch typ {
	case "string", "ip":
		val, err = data.Get(key).String()
	case "float":
		val, err = data.Get(key).Float64()
	case "int", "money":
		val, err = data.Get(key).Int()
	case "base_file":
		var fullFilename string
		fullFilename, err = data.Get(key).String()
		if err != nil {
			return
		}
		val = filepath.Base(fullFilename)
	default:
		panic("invalid key type: " + typ)
	}

	return
}

// Extract values of json according config keys
func (this *AlsParser) valuesOfJsonKeys(data *json.Json) (values []interface{}, indexJson *json.Json, err error) {
	indexJson, _ = json.NewJson([]byte("{}"))

	var currency string
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
			err = errors.New("not found")
			return
		}

		if !key.Visible {
			continue
		}

		if key.Ignores != nil && key.MsgIgnored(val.(string)) {
			err = errors.New("ignored")
			return
		}

		if key.Regex != nil {
			for _, regex := range key.Regex {
				switch regex {
				case "digit":
					val = this.normalizeDigit(val.(string))
				case "token":
					val = this.normalizeBatchToken(val.(string))
				}
			}
		}

		if strings.HasSuffix(key.Name, "currency") {
			currency = val.(string)
		}
		if key.Type == "money" && currency != "" { // currency key必须在money之前定义
			money := float32(val.(int)) * CURRENCY_TABLE[currency]
			val = int(money) / 100 // 以分为单位，而不是元
		}

		values = append(values, val)
		if key.Name == "type" {
			// 'type' is reserved in ElasticSearch
			key.Name = "typ"
		}
		indexJson.Set(key.Name, val)
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
	msg := fmt.Sprintf(format, args...)
	if background {
		logger.Println(msg)
	} else {
		fmt.Println(this.color + msg + COLOR_MAP["Reset"])
	}
}

func (this *AlsParser) blinkColorPrintfLn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if background {
		logger.Println(msg)
	} else {
		fmt.Println(this.color + COLOR_MAP["Blink"] + msg + COLOR_MAP["Reset"])
	}
}

func (this *AlsParser) alarmf(format string, args ...interface{}) {
	msg := fmt.Sprintf("%s", fmt.Sprintf(format, args...))
	if !strings.HasPrefix(msg, this.conf.Title) {
		msg = fmt.Sprintf("%10s %s", this.conf.Title, msg)
	}

	this.chDownstreamAlarm <- msg
}

func (this *AlsParser) beep() {
	fmt.Fprint(os.Stderr, BEEP)
}

func (this *AlsParser) checkError(err error) {
	if err != nil {
		panic(this.id() + ": " + err.Error())
	}
}
