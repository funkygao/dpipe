package engine

import (
	"sync"
)

func LaunchEngine(config *PipelineConfig) {
	log.Println("Starting hekad...")

	var outputsWg sync.WaitGroup
	var err error

	globals := Globals()
	sigChan := make(chan os.Signal)
	globals.sigChan = sigChan

	for name, output := range config.OutputRunners {
		outputsWg.Add(1)
		if err = output.Start(config, &outputsWg); err != nil {
			log.Printf("Output '%s' failed to start: %s", name, err)
			outputsWg.Done()
			continue
		}
		log.Println("Output started: ", name)
	}

	for name, filter := range config.FilterRunners {
		config.filtersWg.Add(1)
		if err = filter.Start(config, &config.filtersWg); err != nil {
			log.Printf("Filter '%s' failed to start: %s", name, err)
			config.filtersWg.Done()
			continue
		}
		log.Println("Filter started: ", name)
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

	for name, input := range config.InputRunners {
		config.inputsWg.Add(1)
		if err = input.Start(config, &config.inputsWg); err != nil {
			log.Printf("Input '%s' failed to start: %s", name, err)
			config.inputsWg.Done()
			continue
		}
		log.Printf("Input started: %s\n", name)
	}

	// wait for sigint
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP, SIGUSR1)

	for !globals.Stopping {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Println("Reload initiated.")
				if err := notify.Post(RELOAD, nil); err != nil {
					log.Println("Error sending reload event: ", err)
				}
			case syscall.SIGINT:
				log.Println("Shutdown initiated.")
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
	config.inputsWg.Wait()

	log.Println("Waiting for decoders shutdown")
	for _, decoder := range config.allDecoders {
		close(decoder.InChan())
		log.Printf("Stop message sent to decoder '%s'", decoder.Name())
	}
	config.decodersWg.Wait()
	log.Println("Decoders shutdown complete")

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
	config.filtersWg.Wait()

	for _, output := range config.OutputRunners {
		config.router.RemoveOutputMatcher() <- output.MatchRunner()
		log.Printf("Stop message sent to output '%s'", output.Name())
	}
	outputsWg.Wait()
	log.Println("Shutdown complete.")
}
