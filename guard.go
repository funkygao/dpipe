/*
            main

    log1    log2    ...     logN

*/
package main

import (
    "bufio"
    "github.com/funkygao/alser/parser"
	"github.com/funkygao/gofmt"
    "io"
    "os"
    "path/filepath"
	"runtime"
	"time"
)

func guard(jsonConfig jsonConfig) {
    parser.SetLogger(logger)
    parser.SetVerbose(options.verbose)
    parser.SetDebug(options.debug)

	if options.tick > 0 {
		ticker = time.NewTicker(time.Second * time.Duration(options.tick))
		go runTicker()
	}

    for _, item := range jsonConfig {
        paths, err := filepath.Glob(item.Pattern)
        if err != nil {
            panic(err)
        }

        for _, logfile := range paths {
            if options.verbose {
                logger.Printf("%s %v", logfile, item.Parsers)
            }

            file, err := os.Open(logfile)
            if err != nil && err != os.ErrExist {
                panic(err)
            }
            defer file.Close()

            reader := bufio.NewReader(file)
            for {
                line, _, err := reader.ReadLine()
                if err != nil {
                    if err == io.EOF {
                        break
                    } else {
                        panic(err)
                    }
                }

                for _, p := range item.Parsers {
                    parser.Dispatch(p, string(line))
                }
            }
        }
    }

	time.Sleep(time.Second * 100)

}

func runTicker() {
	for _ = range ticker.C {
		ms := new(runtime.MemStats)
		runtime.ReadMemStats(ms)
		logger.Println("goroutine:", runtime.NumGoroutine(), "mem:", gofmt.ByteSize(ms.Alloc))
	}

}
