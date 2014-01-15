package plugins

import (
	"fmt"
	"github.com/funkygao/dpipe/engine"
	"strings"
	"time"
)

func indexName(project *engine.ConfProject, indexPattern string,
	date time.Time) string {
	const (
		YM           = "@ym"
		INDEX_PREFIX = "fun_"
	)

	if strings.HasSuffix(indexPattern, YM) {
		prefix := project.IndexPrefix
		fields := strings.SplitN(indexPattern, YM, 2)
		if fields[0] != "" {
			// e,g. rs@ym
			prefix = fields[0]
		}

		return fmt.Sprintf("%s%s_%d_%02d", INDEX_PREFIX, prefix, date.Year(), int(date.Month()))
	}

	return INDEX_PREFIX + indexPattern
}
