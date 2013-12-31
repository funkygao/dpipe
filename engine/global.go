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
	RELOAD = "reaload"
	STOP   = "stop"
)

var (
	availablePlugins = make(map[string]func() Plugin)
	pluginTypeRegex  = regexp.MustCompile("^.*(Filter|Input|Output)$")

	Globals func() *GlobalConfigStruct
)

// Struct for holding global pipeline config values.
type GlobalConfigStruct struct {
	Debug                 bool
	Verbose               bool
	DryRun                bool
	PoolSize              int
	DecoderPoolSize       int //
	PluginChanSize        int
	MaxMsgLoops           int
	MaxMsgProcessInject   uint
	MaxMsgProcessDuration uint64
	MaxMsgTimerInject     uint
	MaxPackIdle           time.Duration
	Stopping              bool
	BaseDir               string

	Logger  *log.Logger
	sigChan chan os.Signal
}

func (this *GlobalConfigStruct) Shutdown() {
	go func() {
		this.sigChan <- syscall.SIGINT
	}()
}

func DefaultGlobals() *GlobalConfigStruct {
	idle, _ := time.ParseDuration("2m")
	return &GlobalConfigStruct{
		Debug:                 false,
		Verbose:               false,
		DryRun:                false,
		PoolSize:              100,
		DecoderPoolSize:       2,
		PluginChanSize:        50,
		MaxMsgLoops:           4,
		MaxMsgProcessInject:   1,
		MaxMsgProcessDuration: 1000000,
		MaxMsgTimerInject:     10,
		MaxPackIdle:           idle,
		Logger:                log.New(os.Stdout, "", log.Ldate|log.Lshortfile|log.Ltime),
	}
}
