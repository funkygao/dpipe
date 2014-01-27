package main

import (
	"fmt"
	"github.com/bmizerany/perks/quantile"
	"github.com/funkygao/als"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	targets = []float64{0.5, 0.66, 0.75, 0.8, 0.9, 0.95, 0.98, 0.99, 1.}
	quants  = make(map[string]*quantile.Stream)
	areas   = []string{
		"ae",
		"us",
		"fr",
		"de",
		"th",
		"nl",
	}
)

type worker struct {
	wg *sync.WaitGroup
}

func init() {
	for _, area := range areas {
		quants[area] = quantile.NewTargeted(targets...)
	}
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
		line    []byte
		err     error
		present bool
		elapsed interface{}
		msg     = als.NewAlsMessage()
	)

	for {
		line, err = reader.ReadLine()
		if err != nil {
			break
		} else {
			msg.FromLine(string(line))
			elapsed, err = msg.FieldValue("_log_info.elapsed", als.KEY_TYPE_FLOAT)
			if err != nil {
				continue
			}

			if _, present = quants[msg.Area]; present {
				quants[msg.Area].Insert(elapsed.(float64))
			}
		}

	}

}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s logfile\n", os.Args[0])
		os.Exit(1)
	}

	logfiles, err := filepath.Glob(os.Args[1] + "*")
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

	// render final report by area
	for area, q := range quants {
		fmt.Printf("%s, total:%d\n", area, q.Count())
		fmt.Println(strings.Repeat("-", 28))
		for _, t := range targets {
			fmt.Printf("Percent %3d: %13.3fms\n", int(100*t), 1000.*q.Query(t))
		}

		fmt.Println()
	}

}
