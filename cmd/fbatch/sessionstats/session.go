package main

import (
	"flag"
	"fmt"
	"github.com/funkygao/als"
	"io"
	"path/filepath"
	"runtime/debug"
)

var (
	// {sid: {"enter":ts1, "exit": ts2, lv: level}}
	sessionsTable map[string]map[string]uint64

	// {sessionLen: {level: N}}
	reportData map[string]map[string]int

	validSessionN   int
	noexitSessionN  int
	invalidSessionN int
	dateStr         string
)

var options struct {
	pattern  string
	area     string
	verbose  bool
	colorize bool
}

func main() {

	debug.SetMaxThreads(6)

	flag.StringVar(&options.pattern, "f", "", "log files to analyze")
	flag.StringVar(&options.area, "a", "", "area")
	flag.BoolVar(&options.verbose, "v", false, "verbose")
	flag.BoolVar(&options.colorize, "c", false, "show num with color")
	flag.Parse()

	sessionsTable = make(map[string]map[string]uint64)
	reportData = make(map[string]map[string]int)

	calcSessionLength(options.pattern, options.area, options.verbose)

	for _, val := range sessionsTable {
		enterAt, enterPresent := val["enter"]
		exitAt, exitPresent := val["exit"]
		if enterPresent && exitPresent {
			if exitAt < enterAt {
				if options.verbose {
					fmt.Println(val, "shit")
				}

				invalidSessionN += 1

				continue
			} else if exitAt == enterAt {
				continue
			}

			lv := als.GroupedLevel(int(val["lv"]))
			sesslen := als.GroupedSessionLen(int(exitAt - enterAt))

			if _, present := reportData[sesslen]; !present {
				reportData[sesslen] = make(map[string]int)
			}

			validSessionN += 1
			reportData[sesslen][lv] += 1
		} else if !exitPresent {
			noexitSessionN += 1
		}

	}

	showReport()

}

func showReport() {
	cells := len(als.SessionLenLabels()) * len(als.LevelLabels())
	for i, sesslen := range als.SessionLenLabels() {
		if i == 0 {
			fmt.Printf("%12s ", "len\\level")
			for _, lv := range als.LevelLabels() {
				fmt.Printf("%9s", lv)
			}
			fmt.Println()
		}

		// print time range
		fmt.Printf("%12s ", sesslen)

		for _, lv := range als.LevelLabels() {
			if !options.colorize {
				fmt.Printf("%9d", reportData[sesslen][lv])
				continue
			}

			if float64(reportData[sesslen][lv])/float64(validSessionN) >= 2.8/float64(cells) {
				fmt.Printf("\x1b[31m%9d\x1b[0m", reportData[sesslen][lv])
			} else {
				fmt.Printf("%9d", reportData[sesslen][lv])
			}
		}

		fmt.Println()
	}

	fmt.Printf("Date %s\n%20s %9d\n%20s %9d\n%20s %9d\n",
		dateStr,
		"valid sessions", validSessionN,
		"no exit sessions", noexitSessionN,
		"invalid sessions", invalidSessionN)
}

func calcSessionLength(pattern, area string, verbose bool) {
	logfiles, err := filepath.Glob(pattern + "*")
	if err != nil {
		panic(err)
	}

	for _, logfile := range logfiles {
		if verbose {
			fmt.Println(logfile, "is being analyzed...")
		}

		reader := als.NewAlsReader(logfile)
		if err := reader.Start(); err != nil {
			panic(err)
		}

		if dateStr == "" {
			dateStr = reader.LogfileYearMonthDate()
		}

	LOOP:
		for {
			line, err := reader.ReadLine()
			switch err {
			case nil:
				feedQuantile(string(line), area, verbose)

			case io.EOF:
				reader.Close()
				break LOOP
			}
		}

	}

}

func feedQuantile(line string, areaFlag string, verboseFlag bool) {
	area, ts, msg, err := als.ParseAlsLine(line)
	if err != nil {
		if verboseFlag {
			fmt.Printf("err[%s] %s\n", err, line)
		}

		return
	}
	if areaFlag != "" && area != areaFlag {
		return
	}

	data, err := als.MsgToJson(msg)
	if err != nil {
		panic(err)
	}

	sid, _ := data.Get("sid").String()
	if sid == "" {
		return
	}
	typ, _ := data.Get("type").String()
	level, _ := data.Get("lv").Int()

	if _, present := sessionsTable[sid]; !present {
		sessionsTable[sid] = make(map[string]uint64)
	}

	// put data into sessionsTable
	sessionsTable[sid][typ] = ts
	sessionsTable[sid]["lv"] = uint64(level)
}
