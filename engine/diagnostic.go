package engine

import (
	"time"
)

// A diagnostic tracker for the pipeline packs pool
type DiagnosticTracker struct {
	// Track all the packs in a pool
	packs []*PipelinePack

	PoolName string
}

func NewDiagnosticTracker(poolName string) *DiagnosticTracker {
	return &DiagnosticTracker{make([]*PipelinePack, 0, 50), poolName}
}

func (this *DiagnosticTracker) AddPack(pack *PipelinePack) {
	this.packs = append(this.packs, pack)
}

func (this *DiagnosticTracker) Run() {
	var (
		pack           *PipelinePack
		earliestAccess time.Time
		pluginCounts   map[PluginRunner]int
		count          int
		runner         PluginRunner
		globals        = Globals()
	)

	idleMax := globals.MaxPackIdle
	probablePacks := make([]*PipelinePack, 0, len(this.packs))
	ticker := time.NewTicker(time.Duration(30) * time.Second)
	defer ticker.Stop()

	if globals.Verbose {
		globals.Printf("Diagnostic[%s] started with 30s\n", this.PoolName)
	}

	for !globals.Stopping {
		<-ticker.C

		probablePacks = probablePacks[:0] // reset
		pluginCounts = make(map[PluginRunner]int)

		// Locate all the packs that have not been touched in idleMax duration
		// that are not recycled
		earliestAccess = time.Now().Add(-idleMax)
		for _, pack = range this.packs {
			if len(pack.diagnostics.pluginRunners) == 0 {
				continue
			}

			if pack.diagnostics.LastAccess.Before(earliestAccess) {
				probablePacks = append(probablePacks, pack)
				for _, runner = range pack.diagnostics.Runners() {
					pluginCounts[runner] += 1
				}
			}
		}

		if len(probablePacks) > 0 {
			globals.Printf("[%s]%d packs have been idle more than %d seconds\n",
				this.PoolName, len(probablePacks), idleMax)
			globals.Printf("[%s]Plugin names and quantities on idle packs:",
				this.PoolName)
			for runner, count = range pluginCounts {
				runner.SetLeakCount(count) // let runner know leak count

				globals.Printf("\t%s: %d", runner.Name(), count)
			}
			globals.Println()
		}
	}
}
