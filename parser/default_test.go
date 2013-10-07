package parser

import (
	"testing"
	"github.com/bmizerany/assert"
	//"fmt"
)

func TestDefaultParserParseLine(t *testing.T) {
	line := `us,1381118458069,{"cheater":10301051,"type":"helpFriendsRewardAction","world_id":"100001823535095","user":"100001823535095","_log_info":{"uid":10301051,"script_id":3183040714,"serial":3,"host":"10.255.8.189","ip":"79.215.100.157"}}`
	p := new(DefaultParser)
	area, ts, data := p.ParseLine(line)
	var (
		exptectedTs  = uint64(1381118458069)
		extectedArea = "us"
	)

	if area != extectedArea {
		t.Error("area: expected", extectedArea, "got", area)
	}
	assert.Equal(t, exptectedTs, ts)
	assert.Equal(t, "helpFriendsRewardAction", data["type"])
	//logInfo := data["_log_info"]
	//fmt.Printf("%#v", data["_log_info"])
	//assert.Equal(t, 3183040714, data["_log_info"]["script_id"])
}
