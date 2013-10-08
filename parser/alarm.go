package parser

import (
	"fmt"
	"time"
)

type Alarm struct {
	Area string
	Host string
	Duration time.Duration
	Info map[string]string
	Count int
}

func (this Alarm) String() string {
	return fmt.Sprintf("%s^%s^%v^%d^%v", this.Area, this.Host, this.Duration, this.Count, this.Info)
}
