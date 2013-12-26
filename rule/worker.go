package rule

import (
	"net/url"
)

// Every worker has 2 modes:
// tail mode and history mode
type ConfWorker struct {
	Enabled     bool     // enabled
	Dsn         string   // data source name base, default file://
	TailGlob    string   // tail_glob
	HistoryGlob string   // history_glob
	Parsers     []string // slice of parser id
}

func (this *ConfWorker) HasParser(parserId string) bool {
	for _, pid := range this.Parsers {
		if pid == parserId {
			return true
		}
	}

	return false
}

func (this *ConfWorker) Scheme() string {
	u, err := url.Parse(this.TailGlob)
	if err != nil {
		panic(err)
	}

	return u.Scheme
}
