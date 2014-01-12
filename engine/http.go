package engine

import (
	"fmt"
	"net/http"
)

func (this *EngineConfig) launchHttpServer() {
	http.HandleFunc("/", this.handleHttpQuery)
	http.ListenAndServe(this.String("http_addr", ":9876"), nil)
}

func (this *EngineConfig) handleHttpQuery(w http.ResponseWriter, r *http.Request) {
	globals := Globals()

	r.ParseForm()
	if globals.Verbose {
		globals.Println(r.Form)
	}

	cmd := r.Form["cmd"][0]
	switch cmd {
	case "runners":
		fmt.Fprintf(w, "filter: %v\noutput: %v", this.FilterRunners, this.OutputRunners)
	case "projects":
		fmt.Fprintf(w, "%v", this.projects)
	case "inputs":
		fmt.Fprintf(w, "%v", this.InputRunners)
	case "router":
		fmt.Fprintf(w, "output: %v\nfilter: %v", this.router.outputMatchers, this.router.filterMatchers)
	default:
		fmt.Fprintf(w, "invalid cmd")
	}
}
