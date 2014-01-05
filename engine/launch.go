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
	globals.Println("Launching Engine...")

	if globals.Verbose {
		globals.Println("Launching Output(s)")
	}
	for name, runner := range e.OutputRunners {
		outputsWg.Add(1)
		if err = runner.Start(e, outputsWg); err != nil {
			outputsWg.Done()
			panic(err)
		}

		if globals.Verbose {
			globals.Printf("Output[%s] started\n", name)
		}
	}

	if globals.Verbose {
		globals.Println("Launching Filter(s)")
	}
	for name, runner := range e.FilterRunners {
		filtersWg.Add(1)
		if err = runner.Start(e, filtersWg); err != nil {
			filtersWg.Done()
			panic(err)
		}

		if globals.Verbose {
			globals.Printf("Filter[%s] started", name)
		}
	}

	// setup the diagnostic trackers
	inputTracker := NewDiagnosticTracker("input")
	injectTracker := NewDiagnosticTracker("inject")

	if globals.Verbose {
		globals.Println("Initializing PipelinePack pools")
	}
	for i := 0; i < globals.PoolSize; i++ {
		inputPack := NewPipelinePack(e.inputRecycleChan)
		inputTracker.AddPack(inputPack)
		e.inputRecycleChan <- inputPack

		injectPack := NewPipelinePack(e.injectRecycleChan)
		injectTracker.AddPack(injectPack)
		e.injectRecycleChan <- injectPack
	}

	go inputTracker.Run()
	go injectTracker.Run()

	e.router.Start()

	if globals.Verbose {
		globals.Println("Launching Input(s)")
	}
	for name, runner := range e.InputRunners {
		inputsWg.Add(1)
		if err = runner.Start(e, inputsWg); err != nil {
			inputsWg.Done()
			panic(err)
		}

		if globals.Verbose {
			globals.Printf("Input[%s] started\n", name)
		}
	}

	globals.sigChan = make(chan os.Signal)

	// now, we have started all runners. next, wait for sigint
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP)

	if globals.Verbose {
		globals.Println("Waiting for os signals...")
	}
	for !globals.Stopping {
		select {
		case sig := <-globals.sigChan:
			globals.Printf("Got signal %s\n", sig.String())
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
		globals.Printf("Stop message sent to '%s'", runner.Name())
	}
	inputsWg.Wait() // wait for all inputs done

	for _, runner := range e.FilterRunners {
		// needed for a clean shutdown without deadlocking or orphaning messages
		// 1. removes the matcher from the router
		// 2. closes the matcher input channel and lets it drain
		// 3. closes the filter input channel and lets it drain
		// 4. exits the filter
		//e.router.RemoveFilterMatcher() <- filter.MatchRunner()
		globals.Printf("Stop message sent to '%s'", runner.Name())
	}
	filtersWg.Wait()

	for _, runner := range e.OutputRunners {
		//e.router.RemoveOutputMatcher() <- output.MatchRunner()
		globals.Printf("Stop message sent to '%s'", runner.Name())
	}
	outputsWg.Wait()

	globals.Println("Shutdown complete.")
}
