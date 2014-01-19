package plugins

import (
	"bytes"
	"fmt"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/bjtime"
	"github.com/funkygao/golib/observer"
	conf "github.com/funkygao/jsconf"
	"sync"
	"time"
)

type AlarmOutput struct {
	// {project: chan}
	emailChans map[string]chan string

	// {project: {camelName: worker}}
	workers map[string]map[string]*alarmWorker
}

func (this *AlarmOutput) Init(config *conf.Conf) {
	this.emailChans = make(map[string]chan string)
	this.workers = make(map[string]map[string]*alarmWorker)

	for i := 0; i < len(config.List("projects", nil)); i++ {
		keyPrefix := fmt.Sprintf("projects[%d].", i)
		proj := config.String(keyPrefix+"name", "")
		this.emailChans[proj] = make(chan string, 20)
		this.workers[proj] = make(map[string]*alarmWorker)

		workersMutex := new(sync.Mutex)

		for j := 0; j < len(config.List(keyPrefix+"workers", nil)); j++ {
			section, err := config.Section(fmt.Sprintf("%sworkers[%d]", keyPrefix, j))
			if err != nil {
				panic(err)
			}

			worker := &alarmWorker{projName: proj, emailChan: this.emailChans[proj],
				workersMutex: workersMutex}
			worker.init(section)
			this.workers[proj][worker.conf.camelName] = worker
		}
	}

}

func (this *AlarmOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		pack       *engine.PipelinePack
		reloadChan = make(chan interface{})
		ok         = true
		inChan     = r.InChan()
	)

	for projName, emailChan := range this.emailChans {
		go this.runSendAlarmsWatchdog(h.Project(projName), emailChan)
	}

	// start all the workers
	goAhead := make(chan bool)
	for _, projectWorkers := range this.workers {
		for _, w := range projectWorkers {
			go w.run(h, goAhead)
			<-goAhead // in case of race condition with worker.inject
		}
	}

	observer.Subscribe(engine.RELOAD, reloadChan)

DONE:
	for ok {
		select {
		case <-reloadChan:
			// TODO

		case pack, ok = <-inChan:
			if !ok {
				break DONE
			}

			this.handlePack(pack)
			pack.Recycle()
		}
	}

	this.stop()

	return nil
}

func (this *AlarmOutput) stop() {
	// stop all the workers
	for _, workers := range this.workers {
		for _, w := range workers {
			w.stop()
		}
	}

	// close alarm email channels
	for _, ch := range this.emailChans {
		close(ch)
	}
}

func (this *AlarmOutput) sendAlarmMailsLoop(project *engine.ConfProject,
	mailBody *bytes.Buffer, bodyLines *int) {
	var (
		globals           = engine.Globals()
		mailConf          = project.MailConf
		mailSleep         = mailConf.SleepStart
		busyLineThreshold = mailConf.BusyLineThreshold
		bodyLineThreshold = mailConf.LineThreshold
		maxSleep          = mailConf.SleepMax
		minSleep          = mailConf.SleepMin
		sleepStep         = mailConf.SleepStep
	)

	for !globals.Stopping {
		select {
		case <-time.After(time.Second * time.Duration(mailSleep)):
			if *bodyLines >= bodyLineThreshold {
				go Sendmail(mailConf.Recipients,
					fmt.Sprintf("ALS[%s] - %d alarms", project.Name, *bodyLines),
					mailBody.String())
				project.Printf("alarm sent=> %s, sleep=%d\n", mailConf.Recipients, mailSleep)

				// backoff sleep
				if *bodyLines >= busyLineThreshold {
					mailSleep -= sleepStep
					if mailSleep < minSleep {
						mailSleep = minSleep
					}
				} else {
					// idle alarm
					mailSleep += sleepStep
					if mailSleep > maxSleep {
						mailSleep = maxSleep
					}
				}

				mailBody.Reset()
				*bodyLines = 0
			}
		}
	}
}

func (this *AlarmOutput) runSendAlarmsWatchdog(project *engine.ConfProject,
	emailChan chan string) {
	var (
		globals   = engine.Globals()
		mailLines int
		mailBody  bytes.Buffer
	)

	go this.sendAlarmMailsLoop(project, &mailBody, &mailLines)

	for line := range emailChan {
		if globals.Debug {
			project.Printf("got email alarm: %s\n", line)
		}

		mailBody.WriteString(fmt.Sprintf("%s %s\n",
			bjtime.TsToString(int(time.Now().UTC().Unix())), line))
		mailLines += 1
	}

}

func (this *AlarmOutput) handlePack(pack *engine.PipelinePack) {
	if worker, present := this.workers[pack.Project][pack.Logfile.CamelCaseName()]; present {
		worker.inject(pack.Message)
	}
}

func init() {
	engine.RegisterPlugin("AlarmOutput", func() engine.Plugin {
		return new(AlarmOutput)
	})
}
