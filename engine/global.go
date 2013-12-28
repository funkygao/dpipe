package worker

import (
	"regexp"
)

var (
	AvailablePlugins = make(map[string]func() interface{})
	PluginTypeRegex  = regexp.MustCompile("^.*(Decoder|Filter|Input|Output)$")
)
