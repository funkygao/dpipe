package main

import (
	"fmt"
	"github.com/funkygao/als"
	"os"
	"path/filepath"
	"sync"
)

type worker struct {
	wg *sync.WaitGroup
}

func (this *worker) run(fn string) {
	reader := als.NewAlsReader(fn)
	if err := reader.Open(); err != nil {
		panic(err)
	}
	defer func() {
		this.wg.Done()
		reader.Close()
	}()

	var (
		line  []byte
		err   error
		snsid interface{}
		msg   = als.NewAlsMessage()
	)

	for {
		line, err = reader.ReadLine()
		if err != nil {
			break
		} else {
			msg.FromLine(string(line))
			snsid, err = msg.FieldValue("_log_info.snsid", als.KEY_TYPE_STRING)
			if err != nil {
				continue
			}

			fetchAvatar(snsid.(string))
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s logfile\n", os.Args[0])
		os.Exit(1)
	}

	logfiles, err := filepath.Glob("/data2/als/pv/*")
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)

	for _, logfile := range logfiles {
		fmt.Printf("[%s]is being analyzed...\n", logfile)
		w := worker{wg: wg}
		wg.Add(1)
		go w.run(logfile)
	}

	// wait all workers finish...
	wg.Wait()

	generateAvatarHtml()

}
