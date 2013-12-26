package main

import (
	"fmt"
	"github.com/funkygao/alser/parser"
	"github.com/funkygao/alser/rule"
	mail "github.com/funkygao/alser/sendmail"
	"github.com/funkygao/gofmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

func runTicker(ticker *time.Ticker, lines *int) {
	startTime := time.Now()
	ms := new(runtime.MemStats)
	for _ = range ticker.C {
		runtime.ReadMemStats(ms)
		logger.Printf("ver:%s, goroutine:%d, mem:%s, workers:%d parsers:%d lines:%d, elapsed:%s\n",
			BuildID,
			runtime.NumGoroutine(), gofmt.ByteSize(ms.Alloc),
			len(allWorkers), parser.ParsersCount(), *lines, time.Since(startTime))
	}
}

func notifyUnGuardedLogs(ruleEngine *rule.RuleEngine) {
	const prefixLen = 3

	if ruleEngine.String("mail.unguarded", "") == "" {
		// disabled
		return
	}

	guardedLogs := make(map[string]bool)
	for _, g := range ruleEngine.Guards {
		var filePrefix string
		if options.tailmode {
			filePrefix = g.TailLogGlob
		} else {
			filePrefix = g.HistoryLogGlob
		}

		baseName := filepath.Base(filePrefix)
		guardedLogs[baseName[:prefixLen]] = true
	}

	// FIXME we assume that all the guarded logs are in the same dir
	var logfile string
	if options.tailmode {
		logfile = ruleEngine.Guards[0].TailLogGlob
	} else {
		logfile = ruleEngine.Guards[0].HistoryLogGlob
	}

	unGuardedLogs := make(map[string]bool)
	baseDir := filepath.Dir(logfile)
	allLogs, _ := filepath.Glob(baseDir + "/*")
	for _, path := range allLogs {
		if stat, _ := os.Stat(path); stat.IsDir() {
			// skip sub directories
			continue
		}

		baseName := filepath.Base(path)
		if _, present := guardedLogs[baseName[:prefixLen]]; !present {
			unGuardedLogs[path] = true
		}
	}

	if len(unGuardedLogs) > 0 {
		files := make([]string, 0)
		for logfile, _ := range unGuardedLogs {
			files = append(files, logfile)
		}
		sort.Strings(files)

		var subject = fmt.Sprintf("ALS Logs Unguarded - %d", len(files))
		var mailBody = ""
		for _, logfile := range files {
			mailBody += logfile + "\n"
		}

		mailTo := ruleEngine.String("unguarded.mail_to", "")
		if err := mail.Sendmail(ruleEngine.String("mail.unguarded", ""), subject, mailBody); err == nil && options.verbose {
			logger.Printf("unguarded logs alarm sent => %s\n", mailTo)
		}
	}

}
