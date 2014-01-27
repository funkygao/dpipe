package main

import (
	"flag"
	"fmt"
	"github.com/funkygao/als"
	"log"
	"path/filepath"
	"sync"
	"time"
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
	flag.StringVar(&srcDir, "s", "/data2/als/pv/", "source logfile dir")
	flag.StringVar(&targetDir, "d", "var/", "target dir")
	flag.Parse()

	logfiles, err := filepath.Glob(srcDir + "*")
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for _ = range ticker.C {
			log.Printf("got avatars: %d", len(allUsers))
		}
	}()

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
