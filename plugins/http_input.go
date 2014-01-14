package plugins

import (
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"net/http"
)

type HttpInput struct {
	ident string
	addr  string

	stopping bool
	runner   engine.InputRunner
}

func (this *HttpInput) Init(config *conf.Conf) {
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
	this.addr = config.String("addr", ":9786")
}

func (this *HttpInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	this.runner = r
	http.HandleFunc("/", this.handleHttpInput)
	err := http.ListenAndServe(this.addr, nil)
	if err != nil {
		return err
	}

	return nil
}

func (this *HttpInput) Stop() {
	this.stopping = true
}

func (this *HttpInput) handleHttpInput(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	pack := <-this.runner.InChan()
	this.runner.Inject(pack)
}

func init() {
	engine.RegisterPlugin("HttpInput", func() engine.Plugin {
		return new(HttpInput)
	})
}
