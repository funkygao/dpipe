package parser

import (
    "strings"
    "time"
)

type histEntry struct {
    sum int64 //
    duration time.Duration
}

// a complex KV data structure
// key can be mixed up for sematics
type hist map[string] histEntry

func newHist() hist {
    return make(map[string] histEntry)
}

func (this hist) Key(args ...string) string {
    const keySeperator = ":"
    return strings.Join(args, keySeperator)
}

func (this hist) Avg() {
}

func (this *hist) Reset() {
}

func (this *hist) Remember(entry histEntry)  {
}

func (this hist) ShouldAlarm() {
}
