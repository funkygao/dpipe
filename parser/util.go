package parser

import (
	json "github.com/bitly/go-simplejson"
	"time"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type logInfo struct {
	uid          int64
	snsid        string
	level        int
	payment_cash int
	uri          string
	scriptId     int64
	serial       int
	host         string
	ip           string
}

// extract _log_info fields from a log data entry
func extractLogInfo(data *json.Json) logInfo {
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

// timestamp of UTC to beijing time
func tsToString(ts int) string {
	t := time.Unix(int64(ts), 0)
	return t.In(tzAjust).Format("01-02 15:04:05")
}
