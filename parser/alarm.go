package parser

import (
	"time"
)

type Alarm struct {
	Area string
	Duration time.Duration
	Info map[string]string
}
