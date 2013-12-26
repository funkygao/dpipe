package parser

import (
	"github.com/bmizerany/assert"
	"github.com/funkygao/alser/rule"
	"testing"
	"time"
)

func TestAlsParserParseLine(t *testing.T) {
	ruleEngine, err := rule.LoadRuleEngine("../etc/alser.cf")
	if err != nil || ruleEngine == nil {
		panic(err)
	}
	line := `us,1381118458069,{"cheater":10301051,"type":"helpFriendsRewardAction","world_id":"100001823535095","user":"100001823535095","_log_info":{"uid":10301051,"script_id":3183040714,"serial":3,"host":"10.255.8.189","ip":"79.215.100.157"}}`
	p := new(AlsParser)
	p.conf = ruleEngine.ParserById("Dau")
	area, ts, msg := p.ParseLine(line)
	data, _ := p.msgToJson(msg)
	var (
		exptectedTs  = uint64(1381118458069 / 1000)
		extectedArea = "us"
	)

	if area != extectedArea {
		t.Error("area: expected", extectedArea, "got", area)
	}
	assert.Equal(t, exptectedTs, ts)

	typ, err := data.Get("type").String()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "helpFriendsRewardAction", typ)
	var expectedScriptId int64 = 3183040714
	scriptId, err := data.Get("_log_info").Get("script_id").Int64()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expectedScriptId, scriptId)
}

func TestPhperrorRegexp(t *testing.T) {
	line := `[10-23 06:38:45] E_WARNING: Illegal offset type in unset  - /mnt/royal_release/6263dab3edbf6675eb9022fa0c32e020/code/application/models/map.php [2828],172.31.25.74`
	matches := phpErrorRegexp.FindAllStringSubmatch(line, 10000)[0]
	host, level, src, msg := matches[6], matches[2], matches[4], matches[3]
	assert.Equal(t, "172.31.25.74", host)
	assert.Equal(t, "E_WARNING", level)
	assert.Equal(t, "/mnt/royal_release/6263dab3edbf6675eb9022fa0c32e020/code/application/models/map.php", src)
	assert.Equal(t, "Illegal offset type in unset ", msg)

	line = `[10-24 01:40:33] E_NOTICE: Undefined variable: buy_cost  - /sgn/htdocs/qa_de/application/controllers/map.php [313],10.245.23.137`
	matches = phpErrorRegexp.FindAllStringSubmatch(line, 10000)[0]
	assert.Equal(t, "E_NOTICE", matches[2])
	assert.Equal(t, "Undefined variable: buy_cost ", matches[3])

	line = `[10-24 01:37:00] E_NOTICE: Undefined index: incProduct  - /sgn/htdocs/qa_de/application/models/map.php [2047],10.245.23.137`
	matches = phpErrorRegexp.FindAllStringSubmatch(line, 10000)[0]
	assert.Equal(t, "E_NOTICE", matches[2])
	assert.Equal(t, "Undefined index: incProduct ", matches[3])
}

func TestIndexEntryNormalizedIndexName(t *testing.T) {
	date := time.Unix(int64(1387274365), 0) // 2013-12-17 17:59:25 +0800 CST

	e := indexEntry{indexName: "ab", typ: "test", date: &date}
	assert.Equal(t, "fun_ab", e.normalizedIndexName())

	e = indexEntry{indexName: "@ym", date: &date}
	assert.Equal(t, "", e.normalizedIndexName()) // failed index name

	e = indexEntry{indexName: "haha@ym", date: &date}
	assert.Equal(t, "fun_haha_2013_12", e.normalizedIndexName())
}
