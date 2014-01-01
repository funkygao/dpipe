package engine

import (
	"github.com/funkygao/golib/observer"
	"log"
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
		log       *log.Logger

		err error
	)

	globals := Globals()
	log = globals.Logger
	log.Println("Launching engine...")

	globals.sigChan = make(chan os.Signal)

	for name, runner := range e.OutputRunners {
		outputsWg.Add(1)
		if err = runner.Start(e, outputsWg); err != nil {
			outputsWg.Done()
			panic(err)
		}

		log.Printf("Output[%s] started\n", name)
	}

	for name, runner := range e.FilterRunners {
		filtersWg.Add(1)
		if err = runner.Start(e, filtersWg); err != nil {
			filtersWg.Done()
			panic(err)
		}

		log.Printf("Filter[%s] started", name)
	}

	// Initialize all of the PipelinePack pools
	for i := 0; i < globals.PoolSize; i++ {
		inputPack := NewPipelinePack(e.inputRecycleChan)
		inputTracker.AddPack(inputPack)
		e.inputRecycleChan <- inputPack

		injectPack := NewPipelinePack(e.injectRecycleChan)
		injectTracker.AddPack(injectPack)
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

		log.Printf("Input[%s] started\n", name)
	}

	// now, we have started all runners. next, wait for sigint
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP)

	for !globals.Stopping {
		select {
		case sig := <-e.sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Println("Reloading...")
				observer.Publish(RELOAD, nil)

			case syscall.SIGINT:
				log.Println("Shutdown...")
				globals.Stopping = true
			}
		}
	}

	// cleanup after shutdown

	for _, input := range e.InputRunners {
		input.Input().Stop()
		log.Printf("Stop message sent to input '%s'", input.Name())
	}
	inputsWg.Wait() // wait for all inputs done

	for _, filter := range e.FilterRunners {
		// needed for a clean shutdown without deadlocking or orphaning messages
		// 1. removes the matcher from the router
		// 2. closes the matcher input channel and lets it drain
		// 3. closes the filter input channel and lets it drain
		// 4. exits the filter
		e.router.RemoveFilterMatcher() <- filter.MatchRunner()
		log.Printf("Stop message sent to filter '%s'", filter.Name())
	}
	filtersWg.Wait()

	for _, output := range e.OutputRunners {
		e.router.RemoveOutputMatcher() <- output.MatchRunner()
		log.Printf("Stop message sent to output '%s'", output.Name())
	}
	outputsWg.Wait()

	log.Println("Shutdown complete.")
}
