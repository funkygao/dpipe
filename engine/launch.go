package engine

import (
	"github.com/funkygao/golib/observer"
	"os"
	"sync"
)

// Start all runners and listens for signals
func LaunchEngine(config *PipelineConfig) {
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

	for name, runner := range config.OutputRunners {
		outputsWg.Add(1)
		if err = runner.Start(config, outputsWg); err != nil {
			panic(err)
		}

		log.Printf("Output[%s] started\n", name)
	}

	for name, runner := range config.FilterRunners {
		filtersWg.Add(1)
		if err = runner.Start(config, filtersWg); err != nil {
			panic(err)
		}

		log.Printf("Filter[%s] started", name)
	}

	// Setup the diagnostic trackers
	inputTracker := NewDiagnosticTracker("input")
	injectTracker := NewDiagnosticTracker("inject")

	// Create the report pipeline pack
	config.reportRecycleChan <- NewPipelinePack(config.reportRecycleChan)

	// Initialize all of the PipelinePacks that we'll need
	for i := 0; i < Globals().PoolSize; i++ {
		inputPack := NewPipelinePack(config.inputRecycleChan)
		inputTracker.AddPack(inputPack)
		config.inputRecycleChan <- inputPack

		injectPack := NewPipelinePack(config.injectRecycleChan)
		injectTracker.AddPack(injectPack)
		config.injectRecycleChan <- injectPack
	}

	go inputTracker.Run()
	go injectTracker.Run()
	config.router.Start()

	for name, runner := range config.InputRunners {
		inputsWg.Add(1)
		if err = runner.Start(config, inputsWg); err != nil {
			panic(err)
		}

		log.Printf("Input[%s] started\n", name)
	}

	// wait for sigint
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGUSR1)

	for !globals.Stopping {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Println("Reloading...")
				observer.Publish(RELOAD, nil)

			case syscall.SIGINT:
				log.Println("Shutdown...")
				globals.Stopping = true

			case SIGUSR1:
				log.Println("Queue report initiated.")
				go config.allReportsStdout()
			}
		}
	}

	config.inputsLock.Lock()
	for _, input := range config.InputRunners {
		input.Input().Stop()
		log.Printf("Stop message sent to input '%s'", input.Name())
	}
	config.inputsLock.Unlock()
	inputsWg.Wait()

	config.filtersLock.Lock()
	for _, filter := range config.FilterRunners {
		// needed for a clean shutdown without deadlocking or orphaning messages
		// 1. removes the matcher from the router
		// 2. closes the matcher input channel and lets it drain
		// 3. closes the filter input channel and lets it drain
		// 4. exits the filter
		config.router.RemoveFilterMatcher() <- filter.MatchRunner()
		log.Printf("Stop message sent to filter '%s'", filter.Name())
	}
	config.filtersLock.Unlock()
	filtersWg.Wait()

	for _, output := range config.OutputRunners {
		config.router.RemoveOutputMatcher() <- output.MatchRunner()
		log.Printf("Stop message sent to output '%s'", output.Name())
	}
	outputsWg.Wait()
	log.Println("Shutdown complete.")
}
