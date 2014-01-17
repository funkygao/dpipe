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
	ticker := time.NewTicker(time.Duration(globals.DiagnosticInterval) * time.Second)
	defer ticker.Stop()

	if globals.Verbose {
		globals.Printf("Diagnostic[%s] started with %ds\n", this.PoolName,
			globals.DiagnosticInterval)
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
			globals.Printf("[%s]%d packs have been idle more than %.0f seconds",
				this.PoolName, len(probablePacks), idleMax.Seconds())
			for runner, count = range pluginCounts {
				runner.setLeakCount(count) // let runner know leak count

				globals.Printf("\t%s: %d", runner.Name(), count)
			}
		}
	}
}
