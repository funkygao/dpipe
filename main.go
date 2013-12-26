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
	ruleEngine, err := rule.LoadRuleEngine(options.config)
	if err != nil || ruleEngine == nil {
		panic(err)
	}

	if options.debug {
		pretty.Logf("%# v\n", ruleEngine.Guards)
		pretty.Logf("%# v\n", ruleEngine.Parsers)
	}

	if options.showparsers {
		fmt.Fprintf(os.Stderr, "All parsers\n%s\n", strings.Repeat("=", 20))
		for _, p := range ruleEngine.Parsers {
			fmt.Fprintf(os.Stderr, "%+v\n", p)
		}
		shutdown()
	}

	if options.parser != "" && !ruleEngine.IsParserApplied(options.parser) {
		fmt.Fprintf(os.Stderr, "Invalid parser: %s\n", options.parser)
		shutdown()
	}

	setupMaxProcsAndProfiler()

	logger.Printf("rule engine[%s] has %d kinds of workers input\n",
		options.config, ruleEngine.CountOfWorkers())

	launch(ruleEngine)

	shutdown()
}
