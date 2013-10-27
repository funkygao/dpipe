package parser

import (
	"fmt"
	json "github.com/bitly/go-simplejson"
	"regexp"
	"time"
)

/*
[09-26 03:33:00] E_WARNING: Illegal offset type in unset  - /mnt/releases/code/code_20130925_rc274/application/models/map.php [2607]
[09-26 03:33:00] E_WARNING: Missing argument 1 for getStaticLastVersion(), called in /mnt/releases/code/code_20130925_rc274/application/models/log.php on line 142 and define
d  - /mnt/releases/code/code_20130925_rc274/application/libraries/functions.php [154]
[09-26 03:33:03] E_WARNING: Missing argument 1 for getStaticLastVersion(), called in /mnt/releases/code/code_20130925_rc274/application/models/log.php on line 142 and define
d  - /mnt/releases/code/code_20130925_rc274/application/libraries/functions.php [154]
[09-26 08:55:04] E_NOTICE: Memcache::get() [<a href='memcache.get'>memcache.get</a>]: Server 10.251.3.167 (tcp 11211) failed with: Connection timed out (110)  - /mnt/releases/code/code_20130925_rc274/system/cache/memcache.php [40]
*/

var (
	// timestamp level: msg - file lineNo,ip
	phpErrorRegexp = regexp.MustCompile(`\[(.+)\] (.+?): (.+) - (.+) \[(.+)\],(.+)`)
)

// Php error log parser
// NOTICE/WARNING/ERROR
type PhpErrorLogParser struct {
	DbParser
}

// Constructor
func newPhpErrorLogParser(name string, chAlarm chan<- Alarm, dbFile, createTable, insertSql string) (parser *PhpErrorLogParser) {
	parser = new(PhpErrorLogParser)
	parser.init(name, chAlarm, dbFile, createTable, insertSql)

	go parser.collectAlarms()

	return
}

func (this *PhpErrorLogParser) ParseLine(line string) (area string, ts uint64, _ *json.Json) {
	var data string
	area, ts, data = this.splitLine(line)

	matches := phpErrorRegexp.FindAllStringSubmatch(data, 10000)[0]
	host, level, src, msg := matches[6], matches[2], matches[4], matches[3]

	this.insert(area, ts, host, level, src, msg)

	return
}

func (this *PhpErrorLogParser) collectAlarms() {
	if dryRun {
		this.chWait <- true
		return
	}

	color := FgYellow
	sleepInterval := time.Duration(this.conf.Int("sleep", 35))
	for {
		time.Sleep(time.Second * sleepInterval)

		this.Lock()
		tsFrom, tsTo, err := this.getCheckpoint("phperror")
		if err != nil {
			this.Unlock()
			continue
		}

		rows := this.query("select count(*) as am, msg, area, host, level from phperror where ts<=? group by msg, area, host order by am desc", tsTo)
		parsersLock.Lock()
		this.logCheckpoint(color, tsFrom, tsTo, "PhpError")
		for rows.Next() {
			var area, msg, host, level string
			var amount int64
			err := rows.Scan(&amount, &msg, &area, &host, &level)
			checkError(err)

			warning := fmt.Sprintf("%5d%3s%12s%16s %s", amount, area, level, host, msg)
			this.colorPrintln(color, warning)
		}
		this.beep()
		parsersLock.Unlock()
		rows.Close()

		if affected := this.execSql("delete from phperror where ts<=?", tsTo); affected > 0 && verbose {
			logger.Printf("phperror %d rows deleted\n", affected)
		}

		this.Unlock()

		if this.stopped {
			this.chWait <- true
			break
		}

	}

}
