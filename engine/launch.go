package engine

import (
	"github.com/funkygao/golib/observer"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Start all runners and listens for signals
func Launch(e *EngineConfig) {
	var (
		outputsWg = new(sync.WaitGroup)
		filtersWg = new(sync.WaitGroup)
		inputsWg  = new(sync.WaitGroup)

		err error
	)

	globals := Globals()
	globals.Println("Launching engine...")

	globals.sigChan = make(chan os.Signal)

	for name, runner := range e.OutputRunners {
		outputsWg.Add(1)
		if err = runner.Start(e, outputsWg); err != nil {
			outputsWg.Done()
			panic(err)
		}

		globals.Printf("Output[%s] started\n", name)
	}
	globals.Println("all Output started")

	for name, runner := range e.FilterRunners {
		filtersWg.Add(1)
		if err = runner.Start(e, filtersWg); err != nil {
			filtersWg.Done()
			panic(err)
		}

		globals.Printf("Filter[%s] started", name)
	}
	globals.Println("all Filter started")

	// Initialize all of the PipelinePack pools
	for i := 0; i < globals.PoolSize; i++ {
		inputPack := NewPipelinePack(e.inputRecycleChan)
		e.inputRecycleChan <- inputPack

		injectPack := NewPipelinePack(e.injectRecycleChan)
		e.injectRecycleChan <- injectPack
	}

	// start the router
	e.router.Start()

	for name, runner := range e.InputRunners {
		inputsWg.Add(1)
		if err = runner.Start(e, inputsWg); err != nil {
			inputsWg.Done()
			panic(err)
		}

		globals.Printf("Input[%s] started\n", name)
	}
	globals.Println("all Input started")

	// now, we have started all runners. next, wait for sigint
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP)

	for !globals.Stopping {
		select {
		case sig := <-globals.sigChan:
			switch sig {
			case syscall.SIGHUP:
				globals.Println("Reloading...")
				observer.Publish(RELOAD, nil)

			case syscall.SIGINT:
				globals.Println("Shutdown...")
				globals.Stopping = true
			}
		}
	}

	// cleanup after shutdown

	for _, project := range e.projects {
		project.Stop()
	}

	for _, runner := range e.InputRunners {
		runner.Input().Stop()
		globals.Printf("Stop message sent to input '%s'", runner.Name())
	}
	inputsWg.Wait() // wait for all inputs done

	for _, runner := range e.FilterRunners {
		// needed for a clean shutdown without deadlocking or orphaning messages
		// 1. removes the matcher from the router
		// 2. closes the matcher input channel and lets it drain
		// 3. closes the filter input channel and lets it drain
		// 4. exits the filter
		//e.router.RemoveFilterMatcher() <- filter.MatchRunner()
		globals.Printf("Stop message sent to filter '%s'", runner.Name())
	}
	filtersWg.Wait()

	for _, runner := range e.OutputRunners {
		//e.router.RemoveOutputMatcher() <- output.MatchRunner()
		globals.Printf("Stop message sent to output '%s'", runner.Name())
	}
	outputsWg.Wait()

	globals.Println("Shutdown complete.")
}
