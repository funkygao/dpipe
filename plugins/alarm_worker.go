package plugins

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/funpipe/engine"
	sqldb "github.com/funkygao/golib/db"
	"github.com/funkygao/golib/stats"
	"github.com/funkygao/gotime"
	conf "github.com/funkygao/jsconf"
	"os"
	"strings"

	"sync"
	"time"
)

type alarmWorkerConfigField struct {
	name    string
	typ     string // float, string(default), int, money
	contain string // only being validator instead of data
	ignores []string
}

func (this *alarmWorkerConfigField) init(config *conf.Conf) {
	this.name = config.String("name", "")
	if this.name == "" {
		panic("alarm worker field name can't be empty")
	}

	this.typ = config.String("type", als.KEY_TYPE_STRING)
	this.contain = config.String("contain", "")
	this.ignores = config.StringList("ignores", nil)
}

func (this *alarmWorkerConfigField) valueOfKey(msg *als.AlsMessage) (value interface{}, err error) {
	value, err = msg.FieldValue(this.name, this.typ)
	return
}

type alarmWorkerConfig struct {
	camelName string
	title     string // optional, defaults to camelName

	fields []alarmWorkerConfigField // besides area,ts

	colors          []string // fg, effects, bg
	printFormat     string   // printf
	instantFormat   string   // instantf, echo for each occurence
	showSummary     bool
	windowSize      time.Duration
	beepThreshold   int
	abnormalPercent float64
	abnormalBase    int

	dbName    string
	tableName string
	persistDb string // will never auto delete for manual analytics

	createTable string
	insertStmt  string
	statsStmt   string
}

func (this *alarmWorkerConfig) init(config *conf.Conf) {
	this.camelName = config.String("camel_name", "")
	if this.camelName == "" {
		panic("empty camel_name in alarmWorkerConfig")
	}

	this.title = config.String("title", "")
	if this.title == "" {
		this.title = this.camelName
	}
	this.colors = config.StringList("colors", nil)
	this.printFormat = config.String("printf", "")
	this.instantFormat = config.String("iprintf", "")
	this.windowSize = time.Duration(config.Int("window_size", 10))
	this.showSummary = config.Bool("show_summary", false)
	this.beepThreshold = config.Int("beep_threshold", 0)
	this.abnormalBase = config.Int("abnormal_base", 10)
	this.abnormalPercent = config.Float("abnormal_percent", 1.5)
	this.dbName = config.String("dbname", "")
	this.tableName = this.dbName
	this.persistDb = config.String("pdbname", "")
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

	db          *sqldb.SqlDb
	pdb         *sqldb.SqlDb // persist db
	insertStmt  *sql.Stmt
	pinsertStmt *sql.Stmt // persist
	statsStmt   *sql.Stmt

	history map[string]int64 // TODO LRU incase of OOM
}

func (this *alarmWorker) init(config *conf.Conf) {
	this.Mutex = new(sync.Mutex)
	this.history = make(map[string]int64)

	this.conf = alarmWorkerConfig{}
	this.conf.init(config)

	this.createDBs()
	this.prepareInsertStmts()
	this.prepareStatsStmt()
}

func (this *alarmWorker) stop() {
	if this.insertStmt != nil {
		this.insertStmt.Close()
	}
	if this.statsStmt != nil {
		this.statsStmt.Close()
	}
	this.db.Close()
	this.pdb.Close()
}

func (this *alarmWorker) run(e *engine.EngineConfig) {
	var (
		globals = engine.Globals()
		summary = stats.Summary{}
	)

	if globals.DryRun {
		return
	}

	this.project = e.Project(this.projName)

	for !globals.Stopping {
		time.Sleep(time.Second * this.conf.windowSize)

		this.Lock()
		windowHead, windowTail, err := this.getCheckpoint()
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
		this.echoCheckpoint(windowHead, windowTail, this.conf.title)
		for rows.Next() {
			for i, _ := range cols {
				valuePtrs[i] = &values[i]
			}

			rows.Scan(valuePtrs...)

			var amount = values[0].(int64)
			if amount == 0 {
				break
			}

			if this.conf.showSummary {
				summary.Add(float64(amount))
			}

			// beep and alarmMail
			if this.conf.beepThreshold > 0 && int(amount) >= this.conf.beepThreshold {
				this.beep()
				this.alarmMail(this.conf.printFormat, values...)
			}

			// abnormal blink
			if amount >= int64(this.conf.abnormalBase) &&
				this.isAbnormalChange(amount, this.historyKey(this.conf.printFormat, values)) {
				this.beep()
				this.blinkColorPrintfLn(this.conf.printFormat, values...)
			}

			this.colorPrintfLn(this.conf.printFormat, values...)
		}

		// show summary
		if this.conf.showSummary && summary.N > 0 {
			this.colorPrintfLn("Total: %.1f, Mean: %.1f", summary.Sum, summary.Mean)
		}

		this.mutex.Unlock()
		rows.Close()

		this.delRecordsBefore(windowTail)
		this.Unlock()
	}

}

func (this *alarmWorker) inject(msg *als.AlsMessage) {
	args, err := this.extractPayloadByFields(msg)
	if err != nil {
		return
	}

	// insert_stmt must be like INSERT INTO (area, ts, ...)
	args = append([]interface{}{msg.Area, msg.Timestamp}, args...)
	this.insert(args...)
}

func (this *alarmWorker) extractPayloadByFields(msg *als.AlsMessage) (values []interface{}, err error) {
	var val interface{}
	values = make([]interface{}, 0, 5)
	for _, field := range this.conf.fields {
		val, err = field.valueOfKey(msg)
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
			this.project.Printf("[%s] history data cleared\n", this.conf.title)
		}

		this.history[key] = amount // refresh
	}()

	if lastAmount, present := this.history[key]; present {
		delta := amount - lastAmount
		if delta < 0 {
			return false
		}

		if float64(delta)/float64(lastAmount) >= this.conf.abnormalPercent {
			return true
		}
	}

	return false
}

func (this *alarmWorker) beep() {
	this.project.Print("\a")
}

func (this *alarmWorker) alarmMail(format string, args ...interface{}) {
	msg := fmt.Sprintf("%s", fmt.Sprintf(format, args...))
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
func (this *alarmWorker) createDBs() {
	const (
		DATA_BASEDIR          = "var"
		SQLITE3_DBFILE_SUFFIX = "sqlite"
		PERSIST_DBFILE_SUFFIX = "db"
	)
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc",
		fmt.Sprintf("%s/%s-%d.%s", DATA_BASEDIR, this.conf.dbName, os.Getpid(),
			SQLITE3_DBFILE_SUFFIX))
	this.db = sqldb.NewSqlDb(sqldb.DRIVER_SQLITE3, dsn, this.project.Logger)
	this.db.SetDebug(engine.Globals().Debug)
	this.db.CreateDb(fmt.Sprintf(this.conf.createTable, this.conf.dbName))

	if this.conf.persistDb != "" {
		dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc",
			fmt.Sprintf("%s/%s.%s", DATA_BASEDIR, this.conf.persistDb,
				PERSIST_DBFILE_SUFFIX))
		this.pdb = sqldb.NewSqlDb(sqldb.DRIVER_SQLITE3, dsn, this.project.Logger)
		this.pdb.SetDebug(engine.Globals().Debug)
		this.pdb.CreateDb(fmt.Sprintf(this.conf.createTable, this.conf.persistDb))
	}
}

func (this *alarmWorker) prepareInsertStmts() {
	if this.conf.insertStmt == "" {
		panic("insert_stmt not configured")
	}

	this.insertStmt = this.db.Prepare(fmt.Sprintf(this.conf.insertStmt, this.conf.dbName))
	if this.pdb != nil {
		this.pinsertStmt = this.pdb.Prepare(
			fmt.Sprintf(this.conf.insertStmt, this.conf.persistDb))
	}
}

func (this *alarmWorker) prepareStatsStmt() {
	statsSql := this.conf.statsSql()
	if statsSql == "" {
		panic("stats_stmt not configured")
	}

	this.statsStmt = this.db.Prepare(statsSql)
}

// auto lock/unlock
func (this *alarmWorker) insert(args ...interface{}) {
	if engine.Globals().Debug {
		this.project.Printf("%s %+v\n", this.conf.title, args)
	}

	this.Lock()
	this.insertStmt.Exec(args...)
	this.Unlock()

	if this.pinsertStmt != nil {
		this.pinsertStmt.Exec(args...)
	}
}

// caller is responsible for locking
func (this *alarmWorker) delRecordsBefore(ts int) (affectedRows int64) {
	affectedRows = this.db.ExecSql("DELETE FROM "+this.conf.dbName+"  WHERE ts<=?", ts)

	return
}

func (this *alarmWorker) getCheckpoint(wheres ...string) (tsFrom, tsTo int, err error) {
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
	err = row.Scan(&tsFrom, &tsTo)
	if err == nil && tsTo == 0 {
		err = errors.New("empty table")
	}

	if engine.Globals().Debug {
		this.project.Println(tsFrom, tsTo, err)
	}

	return
}

func (this *alarmWorker) echoCheckpoint(tsFrom, tsTo int, title string) {
	this.project.Println() // seperator
	this.colorPrintfLn("(%s  ~  %s) %s", gotime.TsToString(tsFrom),
		gotime.TsToString(tsTo), title)
}

func (this *alarmWorker) blinkColorPrintfLn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	this.project.Println(als.Colorize(append(this.conf.colors, "Blink"), msg))
}

func (this *alarmWorker) colorPrintfLn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	this.project.Println(als.Colorize(this.conf.colors, msg))

}
