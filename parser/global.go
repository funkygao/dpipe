package parser

import (
	"log"
	"regexp"
	"sync"
)

var (
	allParsers map[string]Parser = make(map[string]Parser) // registered on init manually
	mutex                        = new(sync.Mutex)         // lock across all parsers

	// passed from main
	logger    *log.Logger
	verbose   bool = false
	debug     bool = false
	dryRun    bool = false
	daemonize bool = false

	beeped int = 1 // how many beeps current proc has been triggered

	chParserAlarm = make(chan string)

	digitsRegexp   = regexp.MustCompile(`\d+`)
	tokenRegexp    = regexp.MustCompile(`pre: .*; current: .*`)
	phpErrorRegexp = regexp.MustCompile(`\[(.+)\] (.+?): (.+) - (.+) \[(.+)\],(.+)`)

	CURRENCY_TABLE = map[string]float32{
		"IDR": 0.00009,
		"VND": 0.000047,
		"NZD": 0.84,
		"HUF": 0.0045,
		"GBP": 1.6,
		"COP": 0.00053,
		"TRY": 0.5,
		"MXN": 0.078,
		"PHP": 0.023,
		"AUD": 0.94,
		"PLN": 0.32,
		"EUR": 1.35,
		"THB": 0.032,
		"MYR": 0.32,
		"BRL": 0.45,
		"INR": 0.016,
		"CAD": 0.97,
		"SAR": 0.27,
		"VEF": 0.16,
		"ARS": 0.17,
		"CZK": 0.052,
		"DKK": 0.18,
		"USD": 1.0,
	}

	COLOR_MAP = map[string]string{
		// e,g. FgBlack + Blink + BgGreen + "hello" + Reset
		"Reset": "\x1b[0m",

		"Bright":     "\x1b[1m",
		"Dim":        "\x1b[2m",
		"Underscore": "\x1b[4m",
		"Blink":      "\x1b[5m",
		"Reverse":    "\x1b[7m",
		"Hidden":     "\x1b[8m",

		"FgBlack":   "\x1b[30m",
		"FgRed":     "\x1b[31m",
		"FgGreen":   "\x1b[32m",
		"FgYellow":  "\x1b[33m",
		"FgBlue":    "\x1b[34m",
		"FgMagenta": "\x1b[35m",
		"FgCyan":    "\x1b[36m",
		"FgWhite":   "\x1b[37m",

		"BgBlack":   "\x1b[40m",
		"BgRed":     "\x1b[41m",
		"BgGreen":   "\x1b[42m",
		"BgYellow":  "\x1b[43m",
		"BgBlue":    "\x1b[44m",
		"BgMagenta": "\x1b[45m",
		"BgCyan":    "\x1b[46m",
		"BgWhite":   "\x1b[47m",
	}
)

const (
	LINE_SPLITTER        = ","
	LINE_SPLIT_NUM       = 3
	MAX_BEEP_VISUAL_HINT = 70

	SQLITE3_DRIVER = "sqlite3"
)

const (
	// error log minus mongo related error
	ERRLOG_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS error (
	area CHAR(10),
	host CHAR(20),
	ts INT,
	cls VARCHAR(50),
    level CHAR(20),
    msg VARCHAR(200) NULL,
    flash INT
);
`
	ERRLOG_INSERT = "INSERT INTO error(area, ts, cls, level, msg, flash, host) VALUES(?,?,?,?,?,?,?)"

	// mongo  error log
	MONGO_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS mongo (
	area CHAR(10),
	host CHAR(20),
	ts INT,
    level CHAR(20),
    msg VARCHAR(200) NULL,
    flash INT
);
`
	MONGO_INSERT = "INSERT INTO mongo(area, ts, level, msg, flash, host) VALUES(?,?,?,?,?,?)"

	// payment log
	PAYMENT_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS payment (
	area CHAR(10),
	host CHAR(20),
	ts INT,
	type VARCHAR(50),
    uid INT(10) NULL,
    level INT,
    amount INT,
    ref VARCHAR(50) NULL,
    item VARCHAR(40),
    currency VARCHAR(20)
);
`
	PAYMENT_INSERT = "INSERT INTO payment(area, host, ts, type, uid, level, amount, ref, item, currency) VALUES(?,?,?,?,?,?,?,?,?,?)"

	// slowresp log
	SLOWRESP_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS slowresp (
	area CHAR(10),
	host CHAR(20),
	uri VARCHAR(50),
	ts INT,
	req_t INT,
	db_t INT	
);
`
	SLOWRESP_INSERT = `INSERT INTO slowresp(area, host, uri, ts, req_t, db_t) VALUES(?,?,?,?,?,?)`

	// level up log
	LEVELUP_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS levelup (
	area CHAR(10),	
	ts INT,
	fromlevel INT
);
`
	LEVELUP_INSERT = `INSERT INTO levelup(area, ts, fromlevel) VALUES(?,?,?)`

	// php errorlog
	PHPERROR_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS phperror (
	area CHAR(10),	
	ts INT,
	host CHAR(25),
	level CHAR(15),
	src_file VARCHAR(80),
	msg VARCHAR(100)
);
`
	PHPERROR_INSERT = `INSERT INTO phperror(area, ts, host, level, src_file, msg) VALUES(?,?,?,?,?,?)`
)
