package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

func guard(jsonConfig jsonConfig) {
	if options.verbose {
		logger.Printf("parsers: %v\n", jsonConfig.parsers())
	}

	for _, item := range jsonConfig {
		paths, err := filepath.Glob(item.Pattern)
		if err != nil {
			panic(err)
		}

		for _, logfile := range paths {
			if options.verbose {
				logger.Printf("%s", logfile)
			}

			file, err := os.Open(logfile)
			if err != nil {
				panic(err)
			}

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

				println(string(line))
			}
		}
	}

}
