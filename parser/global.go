package parser

import (
	"log"
	"sync"
	"time"
)

var (
	allParsers  map[string]Parser // registered on init manually
	parsersLock = new(sync.Mutex) // lock across all parsers, so that println will not be interlaced
	logger      *log.Logger

	tzAjust, _ = time.LoadLocation(TZ)

	verbose bool = false
	debug   bool = false
	dryRun  bool = false

	beeped int = 1

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
)

const (
	version = "0.1.rc"

	LINE_SPLITTER  = ","
	LINE_SPLIT_NUM = 3

	SQLITE3_DRIVER = "sqlite3"
	TZ             = "Asia/Shanghai"

	CONF_DIR = "conf/"
)

// Pass through logger
func SetLogger(l *log.Logger) {
	logger = l
}

// Enable/disable debug mode
func SetDebug(d bool) {
	debug = d
}

// Enable verbose or not
func SetVerbose(v bool) {
	verbose = v
}

func SetDryRun(dr bool) {
	dryRun = dr
}
