package main

import (
    "bufio"
    "github.com/funkygao/alser/parser"
    "io"
    "os"
    "path/filepath"
)

func guard(jsonConfig jsonConfig) {
    parser.SetLogger(logger)
    parser.SetVerbose(options.verbose)
    parser.SetDebug(options.debug)

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

}
