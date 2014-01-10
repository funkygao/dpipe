package benchmark

import (
	"encoding/json"
	"github.com/vmihailenco/msgpack"
	"testing"
)

var in = map[string]interface{}{"c": "LOCK", "k": "31uEbMgunupShBVTewXjtqbBv5MndwfXhb", "T/O": 1000, "max": 200}

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
