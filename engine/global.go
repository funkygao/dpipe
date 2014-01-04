package engine

import (
	"log"
	"os"
	"regexp"
	"syscall"
	"time"
)

const (
	// control channel event types
	RELOAD = "reload"
	STOP   = "stop"
)

var (
	availablePlugins = make(map[string]func() Plugin)
	pluginTypeRegex  = regexp.MustCompile("^.*(Filter|Input|Output)$")

	Globals func() *GlobalConfigStruct
)

// Struct for holding global pipeline config values.
type GlobalConfigStruct struct {
	*log.Logger

	StartedAt      time.Time
	Stopping       bool
	Debug          bool
	Verbose        bool
	DryRun         bool
	Tail           bool
	PoolSize       int
	PluginChanSize int
	TickerLength   int

	MaxMsgLoops           int
	MaxMsgProcessInject   uint
	MaxMsgProcessDuration uint64
	MaxMsgTimerInject     uint
	MaxPackIdle           time.Duration

	sigChan chan os.Signal
}

func (this *GlobalConfigStruct) Shutdown() {
	go func() {
		this.sigChan <- syscall.SIGINT
	}()
}

func (this *GlobalConfigStruct) Debugf(format string, a ...interface{}) {
	this.Printf(format, a...)
}

func DefaultGlobals() *GlobalConfigStruct {
	idle, _ := time.ParseDuration("2m")
	return &GlobalConfigStruct{
		Debug:                 false,
		Verbose:               false,
		DryRun:                false,
		Tail:                  true,
		PoolSize:              100,
		PluginChanSize:        50,
		TickerLength:          10,
		MaxMsgLoops:           4,
		MaxMsgProcessInject:   1,
		MaxMsgProcessDuration: 1000000,
		MaxMsgTimerInject:     10,
		MaxPackIdle:           idle,
		StartedAt:             time.Now(),
		Logger:                log.New(os.Stdout, "", log.Ldate|log.Lshortfile|log.Ltime),
	}
}
