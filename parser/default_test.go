package parser

import (
	"testing"
	//"github.com/bmizerany/assert"
)

func TestParseLine(t *testing.T) {
	line := `us,1381118458069,{"cheater":10301051,"type":"helpFriendsRewardAction","world_id":"100001823535095","user":"100001823535095","_log_info":{"uid":10301051,"script_id":3183040714,"serial":3,"host":"10.255.8.189","ip":"79.215.100.157"}}
us,1381118458094,{"cheater":10301051,"type":"helpFriendsRewardAction","world_id":"100001823535095","user":"100001823535095","_log_info":{"uid":10301051,"script_id":3183040714,"serial":4,"host":"10.255.8.189","ip":"79.215.100.157"}}`
	p := new(DefaultParser)
	p.parseLine(line)


}
