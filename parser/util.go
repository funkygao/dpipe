package parser

import (
	json "github.com/bitly/go-simplejson"
)

func checkError(err error) {
	if err != nil {
		//panic(err)
		logger.Println(err)
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
	info.scriptId, err = infoBody.Get("uid").Int64()
	info.serial, err = infoBody.Get("serial").Int()
	info.host, err = infoBody.Get("host").String()
	info.ip, err = infoBody.Get("ip").String()
	if err != nil {
		// skip error because some column may be null
	}

	return info
}
