package main

import (
	"github.com/funkygao/alser/config"
	"path/filepath"
)

func buddyDataSources(guard config.ConfGuard) []string {
	if guard.IsFileSource() {
		var pattern string
		if options.tailmode {
			pattern = guard.TailLogGlob
		} else {
			pattern = guard.HistoryLogGlob
		}

		logfiles, err := filepath.Glob(pattern)
		if err != nil {
			panic(err)
		}

		return logfiles
	} else if guard.IsDbSource() {

	} else {
		panic("unkown guards data source: " + guard.DataSourceType())
	}

	return nil
}
