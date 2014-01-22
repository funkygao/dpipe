package engine

import (
	"github.com/gorilla/mux"
	"net/http"
)

type PluginHelper interface {
	EngineConfig() *EngineConfig
	PipelinePack(msgLoopCount int) *PipelinePack
	Project(name string) *ConfProject
	RegisterHttpApi(path string,
		handlerFunc func(http.ResponseWriter,
			*http.Request, map[string]interface{}) (interface{}, error)) *mux.Route
}
