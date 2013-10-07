package main

import (
    "fmt"
    "os"
    "runtime"
    "runtime/debug"
)

func init() {
    if _, err := os.Stat(lockfile); err == nil {
        fmt.Fprintf(os.Stderr, "another instance is running, exit\n")
        os.Exit(1)
    }

    file, err := os.Create(lockfile)
    if err != nil {
        panic(err)
    }
    file.Close()

    options = parseFlags()
    options.validate()

    runtime.GOMAXPROCS(runtime.NumCPU()/2 + 1)
}

func main() {
    defer func() {
        cleanup()

        if e := recover(); e != nil {
            debug.PrintStack()
            fmt.Fprintln(os.Stderr, e)
        }
    }()

    logger = newLogger(options)
    logger.Println("starting...")

    jsonConfig := loadConfig(options.config)
    logger.Printf("json config has %d items to guard\n", len(jsonConfig))

    guard(jsonConfig)

    logger.Println("terminated")
}
