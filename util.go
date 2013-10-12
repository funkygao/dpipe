package main

import (
	"os"
)

func instanceLocked() bool {
	_, err := os.Stat(lockfile)
	return err == nil
}

func lockInstance() {
	file, err := os.Create(lockfile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
}
