package main

import (
	"fmt"
	"github.com/funkygao/alser/rule"
	"github.com/kr/pretty"
	"os"
	"runtime/debug"
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

func main() {
	defer func() {
		cleanup()

		if e := recover(); e != nil {
			fmt.Println(e)
			debug.PrintStack()
		}
	}()

	// load the rule engine
	conf, err := config.LoadRuleEngine(options.config)
	if err != nil || conf == nil {
		panic(err)
	}

	if options.debug {
		pretty.Logf("%# v\n", conf.Guards)
		pretty.Logf("%# v\n", conf.Parsers)
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

	setupMaxProcsAndProfiler()

	logger.Printf("conf[%s] has %d kinds of input\n",
		options.config, len(conf.Guards))

	launch(conf)

	shutdown()
}
