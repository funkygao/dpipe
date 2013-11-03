package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"github.conf/funkygao/alser/config"
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

func (this *AlsParser) extractLogInfo(data *json.Json) logInfo {
	info := logInfo{}

	infoBody := data.Get("_log_info")
	info.uid, _ = infoBody.Get("uid").Int64()
	info.snsid, _ = infoBody.Get("snsid").String()
	info.level, _ = infoBody.Get("level").Int()
	info.payment_cash, _ = infoBody.Get("payment_cash").Int()
	info.uri, _ = infoBody.Get("uri").String()
	info.scriptId, _ = infoBody.Get("script_id").Int64()
	info.serial, _ = infoBody.Get("serial").Int()
	info.host, _ = infoBody.Get("host").String()
	info.ip, _ = infoBody.Get("ip").String()

	return info
}

func (this *AlsParser) extractValues() (values []interface{}) {
	values = make([]interface{}, 0)
	var err error
	for _, key := range this.conf.Keys {
		var val interface{}
		switch key.Type {
		case "string":
			val, err = data.Get(key.Key).String()
		case "int":
			val, err = data.Get(key.Key).Int()
		case "float":
			val, err = data.Get(key.Key).Float64()
		}
		if err != nil {
			return
		}

		values = append(values, val)
	}

	if this.conf.LogInfoNeeded() {
		logInfo := this.extractLogInfo(data)
		if this.conf.ShowUri {
			values = append(values, logInfo.uri)
		}
		if this.conf.ShowSrcHost {
			values = append(values, logInfo.host)
		}
	}

	return
}

func (this *AlsParser) Stop() {
}

func (this *AlsParser) Wait() {
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
