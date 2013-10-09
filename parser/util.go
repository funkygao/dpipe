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

	uid := infoBody.Get("uid")
	if uid != nil {
		info.uid, err = uid.Int64()
		checkError(err)
	}

	scriptId := infoBody.Get("script_id")
	if scriptId != nil {
		info.scriptId, err = scriptId.Int64()
		checkError(err)
	}

    serial := infoBody.Get("serial")
    if serial != nil {
        info.serial, err := serial.Int()
        checkError(err)
    }
	
    host := infoBody.Get("host")
    if host != nil {
        info.host, err = host.String()
        checkError(err)
    }
	
    ip := infoBody.Get("ip")
    if ip != nil {
        info.ip, err = ip.String()
        checkError(err)
    }
	
	return info
}
