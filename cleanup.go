package main

import (
    "syscall"
)

func cleanup() {
    syscall.Unlink(lockfile) // cleanup lock file
}
