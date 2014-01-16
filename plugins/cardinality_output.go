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
	counters  *stats.CardinalityCounter
	project   string
	intervals map[string]string
}

func (this *CardinalityOutput) Init(config *conf.Conf) {
	this.counters = stats.NewCardinalityCounter()
	this.project = config.String("project", "")
	this.intervals = make(map[string]string)
}

func (this *CardinalityOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		pack    *engine.PipelinePack
		ok      = true
		project = h.Project(this.project)
		inChan  = r.InChan()
	)

	h.HttpApiHandleFunc("/card/{key}", func(w http.ResponseWriter,
		req *http.Request, params map[string]interface{}) (interface{}, error) {
		return this.handleHttpRequest(w, req, params)
	}).Methods("GET", "PUT")

	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break
			}

			if pack.CardinalityKey != "" && pack.CardinalityData != nil {
				this.intervals[pack.CardinalityKey] = pack.CardinalityInterval

				this.counters.Add(pack.CardinalityKey, pack.CardinalityData)
			}

			pack.Recycle()
		}
	}

	// before we quit, dump counters
	this.dumpCounters(project)

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
		if key != "" {
			output[key] = this.counters.Count(key)
			return output, nil
		}

		// show all counters
		for _, c := range this.counters.Categories() {
			output[c] = fmt.Sprintf("[%v]%d",
				bjtime.TsToString(int(this.counters.StartedAt(c).Unix())),
				this.counters.Count(c))
		}

	case "PUT":
		this.counters.Reset(key)
		output["msg"] = "ok"
	}

	return output, nil
}

func (this *CardinalityOutput) dumpCounters(project *engine.ConfProject) {
	project.Println("Current cardinalities:")
	for _, key := range this.counters.Categories() {
		project.Printf("%15s[%v] %d\n", key,
			bjtime.TsToString(int(this.counters.StartedAt(key).Unix())),
			this.counters.Count(key))
	}
}

func init() {
	engine.RegisterPlugin("CardinalityOutput", func() engine.Plugin {
		return new(CardinalityOutput)
	})
}
