package main

import (
    "os"
    "syscall"
)

func cleanup() {
    syscall.Unlink(lockfile) // cleanup lock file
}

func shutdown() {
    cleanup()
    os.Exit(0)
}
