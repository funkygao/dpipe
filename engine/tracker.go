package engine

type DiagnosticTracker struct {
	// Track all the packs that have been created
	packs []*PipelinePack

	// Identify the name of the recycle channel it monitors packs for
	ChannelName string
}

func NewDiagnosticTracker(channelName string) *DiagnosticTracker {
	return &DiagnosticTracker{make([]*PipelinePack, 0, 50), channelName}
}

// Add a pipeline pack for monitoring
func (this *DiagnosticTracker) AddPack(pack *PipelinePack) {
	this.packs = append(this.packs, pack)
}

// Run the monitoring routine, this should be spun up in a new goroutine
func (d *DiagnosticTracker) Run() {
	var (
		pack           *PipelinePack
		earliestAccess time.Time
		pluginCounts   map[PluginRunner]int
		count          int
		runner         PluginRunner
	)
	g := Globals()
	idleMax := g.MaxPackIdle
	probablePacks := make([]*PipelinePack, 0, len(d.packs))
	ticker := time.NewTicker(time.Duration(30) * time.Second)
	for {
		<-ticker.C
		probablePacks = probablePacks[:0]
		pluginCounts = make(map[PluginRunner]int)

		// Locate all the packs that have not been touched in idleMax duration
		// that are not recycled
		earliestAccess = time.Now().Add(-idleMax)
		for _, pack = range d.packs {
			if len(pack.diagnostics.lastPlugins) == 0 {
				continue
			}
			if pack.diagnostics.LastAccess.Before(earliestAccess) {
				probablePacks = append(probablePacks, pack)
				for _, runner = range pack.diagnostics.Runners() {
					pluginCounts[runner] += 1
				}
			}
		}

		// Drop a warning about how many packs have been idle
		if len(probablePacks) > 0 {
			g.LogMessage("Diagnostics", fmt.Sprintf("%d packs have been idle more than %d seconds.",
				d.ChannelName, len(probablePacks), idleMax))
			g.LogMessage("Diagnostics", fmt.Sprintf("(%s) Plugin names and quantities found on idle packs:",
				d.ChannelName))
			for runner, count = range pluginCounts {
				runner.SetLeakCount(count)
				g.LogMessage("Diagnostics", fmt.Sprintf("\t%s: %d", runner.Name(), count))
			}
			log.Println("")
		}
	}
}
