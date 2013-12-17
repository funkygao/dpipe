package main

import (
	"fmt"
	"github.com/funkygao/alser/rule"
	mail "github.com/funkygao/alser/sendmail"
	"io"
	"log"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"strings"
)

func init() {
	parseFlags()

	if options.showversion {
		showVersion()
	}

	if options.lock {
		if instanceLocked() {
			fmt.Fprintf(os.Stderr, "Another instance is running, exit...\n")
			os.Exit(1)
		}
		lockInstance()
	}

	if options.daemon {
		daemonize(false, true)
	}

	// must be after daemonize, or the pid will be parent pid
	logger = newLogger()

	setupSignals()
}

func newLogger() *log.Logger {
	var logWriter io.Writer = os.Stdout // default log writer
	var err error
	if options.logfile != "" {
		logWriter, err = os.OpenFile(options.logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
	}

	prefix := fmt.Sprintf("[%d]", os.Getpid()) // prefix with pid
	if options.debug {
		return log.New(logWriter, prefix, LOG_OPTIONS_DEBUG)
	}

	return log.New(logWriter, prefix, LOG_OPTIONS)
}

func main() {
	defer func() {
		cleanup()

		if e := recover(); e != nil {
			// console
			fmt.Fprintln(os.Stderr, e)
			debug.PrintStack()

			// log
			logger.Printf("%s\n%s", e, string(debug.Stack()))

			// mail
			mailBody := fmt.Sprintf("%s\n\n%s", e, string(debug.Stack()))
			mail.Sendmail("peng.gao@funplusgame.com", "ALS Crash", mailBody)
		}
	}()

	// load the rule engine
	conf, err := config.LoadRuleEngine(options.config)
	if err != nil || conf == nil {
		panic(err)
	}

	if options.debug {
		logger.Printf("%#v\n", conf.Guards)
		logger.Printf("%#v\n", conf.Parsers)
	}

	if options.showparsers {
		fmt.Fprintf(os.Stderr, "All parsers\n%s\n", strings.Repeat("=", 20))
		for _, p := range conf.Parsers {
			fmt.Fprintf(os.Stderr, "%+v\n", p)
		}
		shutdown()
	}

	if options.parser != "" && !conf.IsParserApplied(options.parser) {
		fmt.Fprintf(os.Stderr, "Invalid parser: %s\n", options.parser)
		shutdown()
	}

	if options.cpuprof != "" {
		f, err := os.Create(options.cpuprof)
		if err != nil {
			panic(err)
		}

		logger.Printf("CPU profiler %s enabled\n", options.cpuprof)
		pprof.StartCPUProfile(f)
	}

	setupMaxProcs()

	logger.Printf("conf[%s] has %d kinds of guards\n",
		options.config, len(conf.Guards))

	guard(conf)

	shutdown()
}
