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

	cmd := r.Form["cmd"]
	fmt.Fprintf(w, "hello from engine")
}
