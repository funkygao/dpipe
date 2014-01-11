package plugins

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/bjtime"
	sqldb "github.com/funkygao/golib/db"
	"github.com/funkygao/golib/stats"
	conf "github.com/funkygao/jsconf"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	errIgnored = errors.New("message ignored")
	errEmpty   = errors.New("empty")

	normalizers = map[string]*regexp.Regexp{
		"digit":       regexp.MustCompile(`\d+`),
		"batch_token": regexp.MustCompile(`pre: .*; current: .*`),
	}
)

type alarmWorkerConfigField struct {
	name        string
	typ         string
	contains    string
	isColumn    bool
	normalizers []string
	ignores     []string
}

func (this *alarmWorkerConfigField) init(config *conf.Conf) {
	this.name = config.String("name", "")
	if this.name == "" {
		panic("alarm worker field name can't be empty")
	}

	this.typ = config.String("type", als.KEY_TYPE_STRING)
	this.contains = config.String("contains", "")
	this.ignores = config.StringList("ignores", nil)
	this.isColumn = config.Bool("is_column", true)
	this.normalizers = config.StringList("normalizers", nil)
}

func (this *alarmWorkerConfigField) value(msg *als.AlsMessage) (val interface{}, err error) {
	val, err = msg.FieldValue(this.name, this.typ)
	if err != nil {
		return
	}

	// contains
	if this.contains != "" {
		if !strings.Contains(val.(string), this.contains) {
			err = errIgnored
			return
		}
	}

	// normalization
	if this.normalizers != nil {
		for _, norm := range this.normalizers {
			normed := normalizers[norm].ReplaceAll([]byte(val.(string)), []byte("?"))
			val = string(normed)
		}
	}

	// ignores
	if this.ignores != nil {
		for _, ignore := range this.ignores {
			if strings.Contains(val.(string), ignore) {
				err = errIgnored
				return
			}
		}
	}

	return
}

type alarmWorkerConfig struct {
	camelName string
	title     string // optional, defaults to camelName

	fields []alarmWorkerConfigField // besides area,ts

	colors          []string // fg, effects, bg
	printFormat     string
	instantFormat   string // 'area' is always 1st col
	showSummary     bool
	windowSize      time.Duration
	beepThreshold   int
	abnormalPercent float64
	abnormalBase    int

	dbName    string
	tableName string

	createTable string
	insertStmt  string
	statsStmt   string
}

func (this *alarmWorkerConfig) init(config *conf.Conf) {
	this.camelName = config.String("camel_name", "")
	if this.camelName == "" {
		panic("empty camel_name")
	}

	this.title = config.String("title", "")
	if this.title == "" {
		this.title = this.camelName
	}
	this.colors = config.StringList("colors", nil)
	this.printFormat = config.String("printf", "")
	this.instantFormat = config.String("iprintf", "")
	this.windowSize = time.Duration(config.Int("window_size", 0)) * time.Second
	this.showSummary = config.Bool("show_summary", false)
	this.beepThreshold = config.Int("beep_threshold", 0)
	this.abnormalBase = config.Int("abnormal_base", 10)
	this.abnormalPercent = config.Float("abnormal_percent", 1.5)
	this.dbName = config.String("dbname", "")
	this.tableName = this.dbName
	this.createTable = config.String("create_table", "")
	this.insertStmt = config.String("insert_stmt", "")
	this.statsStmt = config.String("stats_stmt", "")

	this.fields = make([]alarmWorkerConfigField, 0, 5)
	for i := 0; i < len(config.List("fields", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("fields[%d]", i))
		if err != nil {
			panic(err)
		}

		field := alarmWorkerConfigField{}
		field.init(section)
		this.fields = append(this.fields, field)
	}
}

func (this *alarmWorkerConfig) statsSql() string {
	return fmt.Sprintf(this.statsStmt, this.dbName)
}

type alarmWorker struct {
	*sync.Mutex

	project   *engine.ConfProject
	projName  string
	emailChan chan string
	mutex     *sync.Mutex // accross all alarm workers in a project

	conf alarmWorkerConfig

	db         *sqldb.SqlDb
	insertStmt *sql.Stmt
	statsStmt  *sql.Stmt

	history map[string]int64 // TODO LRU incase of OOM

	instantAlarmOnly bool
}

func (this *alarmWorker) init(config *conf.Conf) {
	this.Mutex = new(sync.Mutex)
	this.history = make(map[string]int64)

	this.conf = alarmWorkerConfig{}
	this.conf.init(config)
	globals := engine.Globals()
	if this.conf.windowSize.Seconds() < 1.0 {
		this.instantAlarmOnly = true

		if this.conf.beepThreshold > 0 {
			globals.Printf("[%s]instant only alarm needn't set 'beep_threshold'", this.conf.camelName)
		}
		if this.conf.abnormalBase > 0 {
			globals.Printf("[%s]instant only alarm needn't set 'abnormal_base'", this.conf.camelName)
		}
		if this.conf.abnormalPercent > 0. {
			globals.Printf("[%s]instant only alarm needn't set 'abnormal_percent'", this.conf.camelName)
		}
		if this.conf.showSummary {
			globals.Printf("[%s]instant only alarm needn't set 'show_summary'", this.conf.camelName)
		}
		if this.conf.createTable != "" {
			globals.Printf("[%s]instant only alarm needn't set 'create_table'", this.conf.camelName)
		}
		if this.conf.statsStmt != "" {
			globals.Printf("[%s]instant only alarm needn't set 'stats_stmt'", this.conf.camelName)
		}
		if this.conf.insertStmt != "" {
			globals.Printf("[%s]instant only alarm needn't set 'insert_stmt'", this.conf.camelName)
		}
		if this.conf.printFormat != "" {
			globals.Printf("[%s]instant only alarm needn't set 'printf'", this.conf.camelName)
		}
		if this.conf.dbName != "" {
			globals.Printf("[%s]instant only alarm needn't set 'dbname'", this.conf.camelName)
		}
	}
}

func (this *alarmWorker) stop() {
	if this.insertStmt != nil {
		this.insertStmt.Close()
	}
	if this.statsStmt != nil {
		this.statsStmt.Close()
	}
	this.db.Close()
}

func (this *alarmWorker) run(h engine.PluginHelper) {
	var (
		globals = engine.Globals()
		summary = stats.Summary{}
		beep    bool
	)

	// lazy assignment
	this.project = h.Project(this.projName)

	if globals.DryRun || this.instantAlarmOnly {
		return
	}

	this.createDB()
	this.prepareInsertStmt()
	this.prepareStatsStmt()

	for !globals.Stopping {
		time.Sleep(this.conf.windowSize)

		this.Lock()
		windowHead, windowTail, err := this.getWindowBorder()
		if err != nil {
			if globals.Verbose {
				this.project.Println(err)
			}

			this.Unlock()
			continue
		}

		if this.conf.showSummary {
			summary.Reset()
		}

		rows, _ := this.statsStmt.Query(windowTail)
		cols, _ := rows.Columns()
		colsN := len(cols)
		values := make([]interface{}, colsN)
		valuePtrs := make([]interface{}, colsN)
		this.mutex.Lock()
		this.printWindowTitle(windowHead, windowTail, this.conf.title)
		for rows.Next() {
			beep = false
			for i, _ := range cols {
				valuePtrs[i] = &values[i]
			}

			rows.Scan(valuePtrs...)

			// 1st column always being aggregated quantile
			var amount = values[0].(int64)
			if amount == 0 {
				break
			}

			if this.conf.showSummary {
				summary.Add(float64(amount))
			}

			// beep and feed alarmMail
			if this.conf.beepThreshold > 0 && int(amount) >= this.conf.beepThreshold {
				beep = true
				this.feedAlarmMail(this.conf.printFormat, values...)
			}

			// abnormal blink
			if amount >= int64(this.conf.abnormalBase) &&
				this.isAbnormalChange(amount, this.historyKey(this.conf.printFormat, values)) {
				this.blinkColorPrintfLn(this.conf.printFormat, values...)
			}

			this.colorPrintfLn(beep, this.conf.printFormat, values...)
		}

		// show summary
		if this.conf.showSummary && summary.N > 0 {
			this.colorPrintfLn(false, "Total: %.1f, Mean: %.1f", summary.Sum, summary.Mean)
		}

		this.mutex.Unlock()
		rows.Close()

		this.moveWindowForward(windowTail)
		this.Unlock()
	}

}

func (this *alarmWorker) inject(msg *als.AlsMessage) {
	args, err := this.fieldValues(msg)
	if err != nil {
		return
	}

	if this.conf.instantFormat != "" {
		iargs := append([]interface{}{area}, args...) // 'area' is always 1st col
		this.colorPrintfLn(true, this.conf.instantFormat, iargs...)
		if this.instantAlarmOnly {
			return
		}
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{msg.Area, msg.Timestamp}, args...)
	this.insert(args...)
}

func (this *alarmWorker) fieldValues(msg *als.AlsMessage) (values []interface{}, err error) {
	var val interface{}
	values = make([]interface{}, 0, 5)

	for _, field := range this.conf.fields {
		if !field.isColumn {
			continue
		}

		val, err = field.value(msg)
		if err != nil {
			return
		}

		values = append(values, val)
	}

	return
}

func (this *alarmWorker) isAbnormalChange(amount int64, key string) bool {
	defer func() {
		// will reset when history size is large enough
		if len(this.history) > (5<<20)/64 {
			// each parser consumes 5M history data
			// each history entry consumes 64bytes
			this.history = make(map[string]int64)
			this.project.Printf("[%s]history data cleared\n", this.conf.title)
		}

		this.history[key] = amount // refresh
	}()

	if lastAmount, present := this.history[key]; present {
		delta := math.Abs(float64(amount - lastAmount))
		if delta/float64(lastAmount) >= this.conf.abnormalPercent {
			return true
		}
	}

	return false
}

func (this *alarmWorker) feedAlarmMail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !strings.HasPrefix(msg, this.conf.title) {
		msg = fmt.Sprintf("%10s %s", this.conf.title, msg)
	}

	this.emailChan <- msg
}

func (this *alarmWorker) historyKey(printf string, values []interface{}) string {
	parts := strings.SplitN(printf, "d", 2) // first column is always amount
	format := strings.TrimSpace(parts[1])
	key := fmt.Sprintf(format, values[1:]...) // raw key

	// use md5 to save memory
	h := md5.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

// create table schema
// for high TPS, each parser has a dedicated sqlite3 db file
func (this *alarmWorker) createDB() {
	const (
		DATA_BASEDIR          = "data"
		SQLITE3_DBFILE_SUFFIX = "sqlite"
	)

	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc",
		fmt.Sprintf("%s/%s-%d.%s", DATA_BASEDIR, this.conf.dbName, os.Getpid(),
			SQLITE3_DBFILE_SUFFIX))
	this.db = sqldb.NewSqlDb(sqldb.DRIVER_SQLITE3, dsn, this.project.Logger)
	this.db.SetDebug(engine.Globals().Debug)
	this.db.CreateDb(fmt.Sprintf(this.conf.createTable, this.conf.dbName))
}

func (this *alarmWorker) prepareInsertStmt() {
	if this.conf.insertStmt == "" {
		panic("insert_stmt not configured")
	}

	this.insertStmt = this.db.Prepare(fmt.Sprintf(this.conf.insertStmt, this.conf.dbName))
}

func (this *alarmWorker) prepareStatsStmt() {
	statsSql := this.conf.statsSql()
	if statsSql == "" {
		panic("stats_stmt not configured")
	}

	this.statsStmt = this.db.Prepare(statsSql)
}

func (this *alarmWorker) insert(args ...interface{}) {
	if engine.Globals().Debug {
		this.project.Printf("%s %+v\n", this.conf.title, args)
	}

	this.Lock()
	this.insertStmt.Exec(args...)
	this.Unlock()
}

// caller is responsible for locking
func (this *alarmWorker) moveWindowForward(tail int) (affectedRows int64) {
	affectedRows = this.db.ExecSql("DELETE FROM "+this.conf.dbName+"  WHERE ts<=?", tail)

	return
}

func (this *alarmWorker) getWindowBorder(wheres ...string) (head, tail int, err error) {
	query := fmt.Sprintf("SELECT min(ts), max(ts) FROM %s", this.conf.dbName)
	if len(wheres) > 0 {
		query += " WHERE 1=1"
		for _, w := range wheres {
			query += " AND " + w
		}
	}

	if engine.Globals().Debug {
		this.project.Println(query)
	}

	row := this.db.QueryRow(query)
	err = row.Scan(&head, &tail)
	if err == nil && tail == 0 {
		err = errEmpty
		return
	}

	if engine.Globals().Debug {
		this.project.Println(head, tail, err)
	}

	return
}

func (this *alarmWorker) printWindowTitle(head, tail int, title string) {
	this.project.Println() // seperator
	this.colorPrintfLn(false, "(%s  ~  %s) %s", bjtime.TsToString(head),
		bjtime.TsToString(tail), title)
}

func (this *alarmWorker) blinkColorPrintfLn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...) + "\a"
	this.project.Println(als.Colorize(append(this.conf.colors, "Blink"), msg))
}

func (this *alarmWorker) colorPrintfLn(beep bool, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if beep {
		msg += "\a"
	}
	this.project.Println(als.Colorize(this.conf.colors, msg))

}
