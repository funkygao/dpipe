package main

import (
	"sync"
)

var (
	allUsers  = map[string]bool{}
	rwMutex   = new(sync.RWMutex)
	present   = false
	targetDir string
	srcDir    string
)
