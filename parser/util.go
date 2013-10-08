package parser

import (
	json "github.com/bitly/go-simplejson"
)

func checkError(err error) {
    if err != nil {
        panic(err)
    }
}

type logInfo struct {
	uid      int64
	scriptId int64
	serial   int
	host     string
	ip       string
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
	//checkError(err)
	return info
}
