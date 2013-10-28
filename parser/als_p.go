package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	conf "github.com/daviddengcn/go-ljson-conf"
	"os"
	"strconv"
	"strings"
)

// Parent parser for all
type AlsParser struct {
	Parser

	name            string
	stopped         bool
	conf            *conf.Conf
	chUpstreamAlarm chan<- Alarm // notify caller
}

func (this *AlsParser) init(name string, chAlarm chan<- Alarm) {
	this.name = name
	this.chUpstreamAlarm = chAlarm
	this.stopped = false

	this.loadConf(CONF_DIR + this.name + ".cf")
}

func (this *AlsParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	var (
		rawData string
		err     error
	)

	area, ts, rawData = this.splitLine(line)

	data, err = json.NewJson([]byte(rawData))
	checkError(err)

	return
}

func (this *AlsParser) Stop() {
	this.stopped = true
}

func (this *AlsParser) Wait() {
}

func (this *AlsParser) colorPrintfLn(color string, format string, args ...interface{}) {
	if daemonize {
		return
	}

	msg := fmt.Sprintf(format, args...)
	fmt.Println(color + msg + Reset)
}

func (this *AlsParser) alarmUpstream(alarm Alarm) {
	this.chUpstreamAlarm <- alarm
}

func (this *AlsParser) alarmParserPrintf(format string, args ...interface{}) {
	if !parserAlarmEnabled {
		return
	}

	msg := fmt.Sprintf(format, args...)
	chParserAlarm <- msg
}

func (this *AlsParser) hasConf() bool {
	return this.conf != nil
}

func (this *AlsParser) loadConf(filename string) {
	var err error
	this.conf, err = conf.Load(filename)
	if err != nil {
		this.conf = nil
	}
}

func (this *AlsParser) beep() {
	fmt.Print("\a")
	if beeped > 80 {
		beeped = 80
	}
	fmt.Fprintln(os.Stderr, strings.Repeat("☼ ", beeped))
	beeped += 1
}

func (this *AlsParser) splitLine(line string) (area string, ts uint64, data string) {
	fields := strings.SplitN(line, LINE_SPLITTER, LINE_SPLIT_NUM)

	area = fields[0]
	var err error
	ts, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}
	ts /= 1000

	data = fields[2]
	return
}
