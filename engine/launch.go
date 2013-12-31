package engine

import (
	"github.com/funkygao/golib/observer"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Start all runners and listens for signals
func LaunchEngine(engine *EngineConfig) {
	var (
		outputsWg = new(sync.WaitGroup)
		filtersWg = new(sync.WaitGroup)
		inputsWg  = new(sync.WaitGroup)

		err error
	)

	globals := Globals()
	var log = globals.Logger
	log.Println("Launching engine...")
	globals.sigChan = make(chan os.Signal)

	for name, runner := range engine.OutputRunners {
		outputsWg.Add(1)
		if err = runner.Start(engine, outputsWg); err != nil {
			panic(err)
		}

		log.Printf("Output[%s] started\n", name)
	}

	for name, runner := range engine.FilterRunners {
		filtersWg.Add(1)
		if err = runner.Start(engine, filtersWg); err != nil {
			panic(err)
		}

		log.Printf("Filter[%s] started", name)
	}

	// Setup the diagnostic trackers
	inputTracker := NewDiagnosticTracker("input")
	injectTracker := NewDiagnosticTracker("inject")

	// Create the report pipeline pack
	engine.reportRecycleChan <- NewPipelinePack(engine.reportRecycleChan)

	// Initialize all of the PipelinePacks that we'll need
	for i := 0; i < globals.PoolSize; i++ {
		inputPack := NewPipelinePack(engine.inputRecycleChan)
		inputTracker.AddPack(inputPack)
		engine.inputRecycleChan <- inputPack

		injectPack := NewPipelinePack(engine.injectRecycleChan)
		injectTracker.AddPack(injectPack)
		engine.injectRecycleChan <- injectPack
	}

	go inputTracker.Run()
	go injectTracker.Run()
	engine.router.Start()

	for name, runner := range engine.InputRunners {
		inputsWg.Add(1)
		if err = runner.Start(engine, inputsWg); err != nil {
			panic(err)
		}

		log.Printf("Input[%s] started\n", name)
	}

	// wait for sigint
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGUSR1)

	for !globals.Stopping {
		select {
		case sig := <-engine.sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Println("Reloading...")
				observer.Publish(RELOAD, nil)

			case syscall.SIGINT:
				log.Println("Shutdown...")
				globals.Stopping = true

			case syscall.SIGUSR1:
				log.Println("Queue report initiated.")
				go engine.allReportsStdout()
			}
		}
	}

	engine.inputsLock.Lock()
	for _, input := range engine.InputRunners {
		input.Input().Stop()
		log.Printf("Stop message sent to input '%s'", input.Name())
	}
	engine.inputsLock.Unlock()
	inputsWg.Wait()

	engine.filtersLock.Lock()
	for _, filter := range engine.FilterRunners {
		// needed for a clean shutdown without deadlocking or orphaning messages
		// 1. removes the matcher from the router
		// 2. closes the matcher input channel and lets it drain
		// 3. closes the filter input channel and lets it drain
		// 4. exits the filter
		engine.router.RemoveFilterMatcher() <- filter.MatchRunner()
		log.Printf("Stop message sent to filter '%s'", filter.Name())
	}
	engine.filtersLock.Unlock()
	filtersWg.Wait()

	for _, output := range engine.OutputRunners {
		engine.router.RemoveOutputMatcher() <- output.MatchRunner()
		log.Printf("Stop message sent to output '%s'", output.Name())
	}
	outputsWg.Wait()
	log.Println("Shutdown complete.")
}
