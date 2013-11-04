package main

import (
	"fmt"
	"github.com/funkygao/alser/config"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"strings"
)

func init() {
	options = parseFlags()

	if options.showversion {
		showVersion()
	}

	logger = newLogger(options) // create logger as soon as possible

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

	setupSignals()
}

func main() {
	defer func() {
		cleanup()

		if e := recover(); e != nil {
			debug.PrintStack()
			fmt.Fprintln(os.Stderr, e)
		}
	}()

	// load the big biz logic config file
	conf, err := config.LoadConfig(options.config)
	if err != nil || conf == nil {
		panic(err)
	}

	if options.showparsers {
		fmt.Fprintf(os.Stderr, "All parsers\n%s\n%+v\n", strings.Repeat("=", 20), conf.Parsers)
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

	logger.Printf("%s has %d kinds of logs to guard\n", options.config, len(conf.Guards))

	guard(conf)

	shutdown()
}
