package benchmark

import (
	"encoding/json"
	"github.com/vmihailenco/msgpack"
	"testing"
)

var in = map[string]interface{}{
	"c":   "LOCK",
	"k":   "31uEbMgunupShBVTewXjtqbBv5MndwfXhb",
	"T/O": 1000,
	"max": 200,
}

const (
	jsonLineForTest = `{"uri":"\/?fb_source=notification&request_ids=629862167081523%2C231759870340420%2C597190080352387%2C640624999328961%2C235464713291862%2C753053901389297%2C790469374302126%2C192819610918125%2C1409213372656992%2C1395677210684824%2C219547141565670%2C445351695593355%2C353291448144469%2C374894915987858%2C1405041129742942%2C1386152901642951%2C1444273795788958%2C268848269934670&ref=notif&app_request_type=user_to_user&notif_t=app_request","_log_info":{"uid":10304512,"snsid":"100006490632784","level":39,"gender":"male","payment_cash":13,"script_id":9524283412,"serial":1,"uri":"\/","host":"172.31.7.194","ip":"81.65.52.251","callee":"POST+\/+44eae87","sid":"adadf","elapsed":0.014667987823486}}`
)

type logInfo struct {
	Callee    string
	Elapsed   float64
	Gender    string
	Host, Ip  string
	Level     int
	P         int `json:"payment_cash"`
	Script_id int `json:"script_id"`
	Sid       string
	Serial    int
	Snsid     string
	Uid       int
	Uri       string
}

type jsonLine struct {
	Uri     string
	Loginfo logInfo `json:"_log_info"`
}

func BenchmarkMultiply(b *testing.B) {
	n := 1
	for i := 0; i < b.N; i++ {
		n *= 2
		if n > (1 << 30) {
			n = 1
		}
	}
}

func BenchmarkJSONEncodeAndDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jsonB := EncodeJSON(in)
		DecodeJSON(jsonB)
	}
}

func BenchmarkMsgPackEncodeAndDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := EncodeMsgPack(in)
		DecodeMsgPack(b)
	}
}

func BenchmarkJsonUnmarshalInterface(b *testing.B) {
	var l interface{}
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(jsonLineForTest), &l)
		//b.Logf("%#v", l)
	}
	b.SetBytes(int64(len([]byte(jsonLineForTest))))
}

func BenchmarkJsonUnmarshalStruct(b *testing.B) {
	var l jsonLine
	for i := 0; i < b.N; i++ {
		json.Unmarshal([]byte(jsonLineForTest), &l)
		//b.Logf("%#v", l)
	}
	b.SetBytes(int64(len([]byte(jsonLineForTest))))
}

func EncodeMsgPack(message map[string]interface{}) []byte {
	b, _ := msgpack.Marshal(message)
	return b
}

func DecodeMsgPack(b []byte) (out map[string]interface{}) {
	_ = msgpack.Unmarshal(b, &out)
	return
}

func EncodeJSON(message map[string]interface{}) []byte {
	b, _ := json.Marshal(message)
	return b
}

func DecodeJSON(b []byte) (out map[string]interface{}) {
	_ = json.Unmarshal(b, &out)
	return
}
