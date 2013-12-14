package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

func instanceLocked() bool {
	_, err := os.Stat(LOCKFILE)
	return err == nil
}

func lockInstance() {
	pid := fmt.Sprintf("%d", os.Getpid())
	if err := ioutil.WriteFile(LOCKFILE, []byte(pid), 0644); err != nil {
		panic(err)
	}
}

func unlockInstance() {
	syscall.Unlink(LOCKFILE)
}
