package main

import (
    . "os"
    "os/signal"
    "strings"
)

func trapSignals() {
    ch := make(chan Signal, 10)
    signal.Notify(ch, caredSignals...)

    go func() {
        sig := <- ch
        for _, s := range caredSignals {
            if s == sig {
                logger.Printf("%s signal recved\n", strings.ToUpper(sig.String()))
				logger.Println("terminated")

                shutdown()
            }
        }
    }()
}

