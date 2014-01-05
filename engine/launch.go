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
		globals.Println("Launching Output(s)...")
	}
	for name, runner := range e.OutputRunners {
		if globals.Verbose {
			globals.Printf("Starting %s\n", name)
		}

		outputsWg.Add(1)
		if err = runner.Start(e, outputsWg); err != nil {
			outputsWg.Done()
			panic(err)
		}
	}

	if globals.Verbose {
		globals.Println("Launching Filter(s)...")
	}
	for name, runner := range e.FilterRunners {
		if globals.Verbose {
			globals.Printf("Starting %s\n", name)
		}

		filtersWg.Add(1)
		if err = runner.Start(e, filtersWg); err != nil {
			filtersWg.Done()
			panic(err)
		}
	}

	// setup the diagnostic trackers
	inputTracker := NewDiagnosticTracker("input")
	injectTracker := NewDiagnosticTracker("inject")

	if globals.Verbose {
		globals.Printf("Initializing PipelinePack pools %d\n", globals.PoolSize)
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
		globals.Println("Launching Input(s)...")
	}
	for name, runner := range e.InputRunners {
		if globals.Verbose {
			globals.Printf("Starting %s\n", name)
		}

		inputsWg.Add(1)
		if err = runner.Start(e, inputsWg); err != nil {
			inputsWg.Done()
			panic(err)
		}

	}

	globals.Println("Engine ready")

	// now, we have started all runners. next, wait for sigint
	globals.sigChan = make(chan os.Signal)
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP)
	for !globals.Stopping {
		select {
		case sig := <-globals.sigChan:
			globals.Printf("Got signal %s\n", sig.String())
			switch sig {
			case syscall.SIGHUP:
				globals.Println("Reloading...")
				observer.Publish(RELOAD, nil)

			case syscall.SIGINT:
				globals.Println("Engine shutdown...")
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

		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}
	}
	inputsWg.Wait() // wait for all inputs done
	if globals.Verbose {
		globals.Println("All Inputs terminated")
	}

	for _, runner := range e.FilterRunners {
		e.router.removeFilterMatcher <- runner.MatchRunner()

		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}
	}
	filtersWg.Wait()
	if globals.Verbose {
		globals.Println("All Filters terminated")
	}

	for _, runner := range e.OutputRunners {
		e.router.removeOutputMatcher <- runner.MatchRunner()

		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}
	}
	outputsWg.Wait()
	if globals.Verbose {
		globals.Println("All Outputs terminated")
	}

	globals.Println("Engine shutdown complete.")
}
