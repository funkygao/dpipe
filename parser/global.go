package parser

import (
	"github.com/abh/geoip"
	"log"
	"regexp"
	"sync"
)

var (
	allParsers map[string]Parser = make(map[string]Parser) // key is parserId
	indexer    *Indexer
	geo        *geoip.GeoIP

	mutex = new(sync.Mutex) // lock across all parsers

	digitsRegexp      = regexp.MustCompile(`\d+`)
	batchTokenRegexp  = regexp.MustCompile(`pre: .*; current: .*`)
	phpErrorRegexp    = regexp.MustCompile(`\[(.+)\] (.+?): (.+) - (.+) \[(.+)\],(.+)`)
	syslogngDropped   = regexp.MustCompile(`dropped=\'program\((.+?)\)=(\d+)\'`)
	syslogngProcessed = regexp.MustCompile(`processed=\'destination\((.+?)\)=(\d+)\'`)

	// passed from main
	logger     *log.Logger
	verbose    bool = false
	debug      bool = false
	dryRun     bool = false
	background bool = false

	chParserAlarm = make(chan string, 10)

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
	DATA_BASEDIR          = "var"
	SQLITE3_DBFILE_SUFFIX = "sqlite"
	PERSIST_DBFILE_SUFFIX = "db"

	LINE_SPLITTER  = ","
	LINE_SPLIT_NUM = 3
	BEEP           = "\a"

	INDEX_YEARMONTH = "@ym"
)

const (
	KEY_TYPE_STRING   = "string" // default type
	KEY_TYPE_IP       = "ip"
	KEY_TYPE_FLOAT    = "float"
	KEY_TYPE_INT      = "int"
	KEY_TYPE_MONEY    = "money"
	KEY_TYPE_BASEFILE = "base_file"

	KEY_NAME_CURRENCY = "currency"
	KEY_NAME_LOCATION = "loc"
	KEY_NAME_TYPE     = "typ"
)
