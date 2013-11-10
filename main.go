package main

import (
	"fmt"
	"github.com/funkygao/alser/config"
	mail "github.com/funkygao/alser/sendmail"
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
	} else if options.verbose {
		logger.Println("instance locking disabled")
	}

	if options.daemon {
		if options.verbose {
			logger.Println("daemonizing...")
		}

		daemonize(false, true)
	}

	setupSignals()
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

	// load the big biz logic config file
	conf, err := config.LoadConfig(options.config)
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

	logger.Printf("conf[%s] has %d kinds of logs to guard\n",
		options.config, len(conf.Guards))

	guard(conf)

	shutdown()
}
