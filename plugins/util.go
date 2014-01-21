package plugins

import (
	"fmt"
	"github.com/funkygao/dpipe/engine"
	"strings"
	"time"
)

func indexName(project *engine.ConfProject, indexPattern string,
	date time.Time) (index string) {
	const (
		YM  = "@ym"
		YMW = "@ymw"
		YMD = "@ymd"

		INDEX_PREFIX = "fun_"
	)

	if strings.Contains(indexPattern, YM) {
		prefix := project.IndexPrefix
		fields := strings.SplitN(indexPattern, YM, 2)
		if fields[0] != "" {
			// e,g. rs@ym
			prefix = fields[0]
		}

		switch indexPattern {
		case YM:
			index = fmt.Sprintf("%s%s_%d_%02d", INDEX_PREFIX, prefix,
				date.Year(), int(date.Month()))
		case YMW:
			year, week := date.ISOWeek()
			index = fmt.Sprintf("%s%s_%d_w%02d", INDEX_PREFIX, prefix,
				year, week)
		case YMD:
			index = fmt.Sprintf("%s%s_%d_%02d_%02d", INDEX_PREFIX, prefix,
				date.Year(), int(date.Month()), date.Day())
		}

		return
	}

	index = INDEX_PREFIX + indexPattern

	return
}
