package engine

import (
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/observer"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func (e *EngineConfig) ServeForever() {
	var (
		outputsWg = new(sync.WaitGroup)
		filtersWg = new(sync.WaitGroup)
		inputsWg  = new(sync.WaitGroup)
		globals   = Globals()
		err       error
	)

	// setup signal handler first to avoid race condition
	// if Input terminates very soon, global.Shutdown will
	// not be able to trap it
	globals.sigChan = make(chan os.Signal)
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP,
		syscall.SIGUSR2, syscall.SIGUSR1)

	e.launchHttpServ()

	if globals.Verbose {
		globals.Println("Launching Output(s)...")
	}
	for _, runner := range e.OutputRunners {
		outputsWg.Add(1)
		if err = runner.start(e, outputsWg); err != nil {
			panic(err)
		}
	}

	if globals.Verbose {
		globals.Println("Launching Filter(s)...")
	}
	for _, runner := range e.FilterRunners {
		filtersWg.Add(1)
		if err = runner.start(e, filtersWg); err != nil {
			panic(err)
		}
	}

	// setup the diagnostic trackers
	inputPackTracker := NewDiagnosticTracker("inputPackTracker")
	e.diagnosticTrackers[inputPackTracker.PoolName] = inputPackTracker
	filterPackTracker := NewDiagnosticTracker("filterPackTracker")
	e.diagnosticTrackers[filterPackTracker.PoolName] = filterPackTracker

	if globals.Verbose {
		globals.Printf("Initializing PipelinePack pools with size=%d\n",
			globals.RecyclePoolSize)
	}
	for i := 0; i < globals.RecyclePoolSize; i++ {
		inputPack := NewPipelinePack(e.inputRecycleChan)
		inputPackTracker.AddPack(inputPack)
		e.inputRecycleChan <- inputPack

		filterPack := NewPipelinePack(e.filterRecycleChan)
		filterPackTracker.AddPack(filterPack)
		e.filterRecycleChan <- filterPack
	}

	go inputPackTracker.Run(e.Int("diagnostic_interval", 20))
	go filterPackTracker.Run(e.Int("diagnostic_interval", 20))

	// check if we have enough recycle pool reservation
	go func() {
		t := time.NewTicker(time.Second * time.Duration(globals.TickerLength))
		defer t.Stop()

		var inputPoolSize, filterPoolSize int

		for _ = range t.C {
			inputPoolSize = len(e.inputRecycleChan)
			filterPoolSize = len(e.filterRecycleChan)
			if globals.Verbose || inputPoolSize == 0 || filterPoolSize == 0 {
				globals.Printf("Recycle poolSize: [input]%d [filter]%d",
					inputPoolSize, filterPoolSize)
			}
		}
	}()

	go e.router.Start()

	for _, project := range e.projects {
		project.Start()
	}

	if globals.Verbose {
		globals.Println("Launching Input(s)...")
	}
	for _, runner := range e.InputRunners {
		inputsWg.Add(1)
		if err = runner.start(e, inputsWg); err != nil {
			inputsWg.Done()
			panic(err)
		}
	}

	globals.Println("Engine mainloop, waiting for signals...")
	go runShutdownWatchdog(e)

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

			case syscall.SIGUSR1:
				observer.Publish(SIGUSR1, nil)

			case syscall.SIGUSR2:
				observer.Publish(SIGUSR2, nil)
			}
		}
	}

	// cleanup after shutdown
	inputPackTracker.Stop()
	filterPackTracker.Stop()

	e.Lock()
	for _, runner := range e.InputRunners {
		if runner == nil {
			// this Input plugin already exit
			continue
		}

		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}

		runner.Input().Stop()
	}
	e.Unlock()
	inputsWg.Wait() // wait for all inputs done
	if globals.Verbose {
		globals.Println("All Inputs terminated")
	}

	// ok, now we are sure no more inputs, but in route.inChan there
	// still may be filter injected packs and output not consumed packs
	// we must wait for all the packs to be consumed before shutdown

	for _, runner := range e.FilterRunners {
		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}

		e.router.removeFilterMatcher <- runner.Matcher()
	}
	filtersWg.Wait()
	if globals.Verbose {
		globals.Println("All Filters terminated")
	}

	for _, runner := range e.OutputRunners {
		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}

		e.router.removeOutputMatcher <- runner.Matcher()
	}
	outputsWg.Wait()
	if globals.Verbose {
		globals.Println("All Outputs terminated")
	}

	//close(e.router.inChan)

	e.stopHttpServ()

	for _, project := range e.projects {
		project.Stop()
	}

	globals.Printf("Shutdown with input:%s, dispatched:%s",
		gofmt.Comma(e.router.stats.TotalInputMsgN),
		gofmt.Comma(e.router.stats.TotalProcessedMsgN))
}
