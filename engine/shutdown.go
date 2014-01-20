package engine

import (
	"time"
)

// Automatically check if no more Input and time to shutdown
func runShutdownWatchdog(e *EngineConfig) {
	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()

	var (
		allInputsDone bool
		globals       = Globals()
	)

	for {
		select {
		case <-ticker.C:
			allInputsDone = true
			for _, runner := range e.InputRunners {
				if runner != nil {
					allInputsDone = false
					break
				}
			}

			if allInputsDone {
				globals.Println("All Input done, shutdown...")
				globals.Shutdown()
				return
			}
		}
	}
}
