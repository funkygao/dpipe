package main

import (
	"os"
)

func instanceLocked() bool {
	_, err := os.Stat(LOCKFILE)
	return err == nil
}

func lockInstance() {
	file, err := os.Create(LOCKFILE)
	if err != nil {
		panic(err)
	}
	defer file.Close()
}
