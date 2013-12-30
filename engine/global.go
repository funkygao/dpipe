package engine

import (
	"os"
	"regexp"
	"syscall"
	"time"
)

const (
	RELOAD = "reaload"
	STOP   = "stop"
)

var (
	AvailablePlugins = make(map[string]func() interface{})
	PluginTypeRegex  = regexp.MustCompile("^.*(Filter|Input|Output)$")
)

// Struct for holding global pipeline config values.
type GlobalConfigStruct struct {
	PoolSize              int
	DecoderPoolSize       int
	PluginChanSize        int
	MaxMsgLoops           uint
	MaxMsgProcessInject   uint
	MaxMsgProcessDuration uint64
	MaxMsgTimerInject     uint
	MaxPackIdle           time.Duration
	Stopping              bool
	BaseDir               string
	sigChan               chan os.Signal
}

func (this *GlobalConfigStruct) Shutdown() {
	go func() {
		this.sigChan <- syscall.SIGINT
	}()
}

var Globals func() *GlobalConfigStruct

func DefaultGlobals() (globals *GlobalConfigStruct) {
	idle, _ := time.ParseDuration("2m")
	return &GlobalConfigStruct{
		PoolSize:              100,
		DecoderPoolSize:       2,
		PluginChanSize:        50,
		MaxMsgLoops:           4,
		MaxMsgProcessInject:   1,
		MaxMsgProcessDuration: 1000000,
		MaxMsgTimerInject:     10,
		MaxPackIdle:           idle,
	}
}
