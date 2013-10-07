package parser

import (
	"strings"
	"time"
	json "github.com/bitly/go-simplejson"
	"strconv"
)

type logInfo struct {
	uid int64
	scriptId int64
	serial int
	host string
	ip string
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// extract _log_info fields from a log data entry
func extractLogInfo(data *json.Json) logInfo {
	var err error
	info := logInfo{}
	infoBody := data.Get("_log_info")
	info.uid, err = infoBody.Get("uid").Int64()
	checkError(err)
	info.scriptId, err = infoBody.Get("script_id").Int64()
	checkError(err)
	info.serial, err = infoBody.Get("serial").Int()
	checkError(err)
	info.host, err = infoBody.Get("host").String()
	checkError(err)
	info.ip, err = infoBody.Get("ip").String()
	checkError(err)
	return info
}

type DefaultParser struct {
}

func (this DefaultParser) ParseLine(line string) (area string, ts uint64, data *json.Json) {
	fields := strings.SplitN(line, LINE_SPLITTER, LINE_SPLIT_NUM)

	area = fields[0]
	var err error
	ts, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		panic(err)
	}

	data, err = json.NewJson([]byte(fields[2]))
	if err != nil {
		panic(err)
	}

	return
}

func (this DefaultParser) GetStats(duration time.Duration) {

}
