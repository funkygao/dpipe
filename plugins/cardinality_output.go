package plugins

import (
	"fmt"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/bjtime"
	"github.com/funkygao/golib/stats"
	conf "github.com/funkygao/jsconf"
	"github.com/gorilla/mux"
	"net/http"
)

type CardinalityOutput struct {
	counters   *stats.CardinalityCounter
	checkpoint string
}

func (this *CardinalityOutput) Init(config *conf.Conf) {
	this.checkpoint = config.String("checkpoint", "")
	this.counters = stats.NewCardinalityCounter()
	if this.checkpoint != "" {
		this.counters.Load(this.checkpoint)
	}
}

func (this *CardinalityOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

	h.RegisterHttpApi("/card/{key}", func(w http.ResponseWriter,
		req *http.Request, params map[string]interface{}) (interface{}, error) {
		return this.handleHttpRequest(w, req, params)
	}).Methods("GET", "PUT")

LOOP:
	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break LOOP
			}

			if pack.CardinalityKey != "" && pack.CardinalityData != nil {
				this.counters.Add(pack.CardinalityKey, pack.CardinalityData)
			}

			pack.Recycle()
		}
	}

	// before we quit, dump counters
	if this.checkpoint != "" {
		this.counters.Dump(this.checkpoint)
	}

	return nil
}

func (this *CardinalityOutput) handleHttpRequest(w http.ResponseWriter,
	req *http.Request, params map[string]interface{}) (interface{}, error) {
	vars := mux.Vars(req)
	key := vars["key"]
	globals := engine.Globals()
	if globals.Verbose {
		globals.Println(req.Method, key)
	}

	output := make(map[string]interface{})
	switch req.Method {
	case "GET":
		if key == "all" {
			for _, c := range this.counters.Categories() {
				output[c] = fmt.Sprintf("[%v] %d",
					bjtime.TsToString(int(this.counters.StartedAt(c).Unix())),
					this.counters.Count(c))
			}
		} else {
			output[key] = this.counters.Count(key)
		}

	case "PUT":
		this.counters.Reset(key)
		output["msg"] = "ok"
	}

	return output, nil
}

func init() {
	engine.RegisterPlugin("CardinalityOutput", func() engine.Plugin {
		return new(CardinalityOutput)
	})
}
