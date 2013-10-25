package parser

import (
	"log"
	"sync"
	"time"
)

var (
	allParsers  map[string]Parser = make(map[string]Parser) // registered on init manually
	parsersLock                   = new(sync.Mutex)         // lock across all parsers, so that println will not be interlaced

	tzAjust, _ = time.LoadLocation(TZ) // same time info for all locales

	logger  *log.Logger // shared with alser
	verbose bool        = false
	debug   bool        = false
	dryRun  bool        = false

	beeped int = 1 // how many beeps that has been triggered

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
	LINE_SPLITTER  = ","
	LINE_SPLIT_NUM = 3

	SQLITE3_DRIVER = "sqlite3"
	TZ             = "Asia/Shanghai"

	CONF_DIR = "conf/"
)
