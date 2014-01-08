package plugins

import (
	"bytes"
	"fmt"
	"github.com/funkygao/funpipe/engine"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/gotime"
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

		for j := 0; j < len(config.List(keyPrefix+"workers", nil)); i++ {
			section, err := config.Section(fmt.Sprintf("%sworkers[%d]", keyPrefix, j))
			if err != nil {
				panic(err)
			}

			worker := &alarmWorker{projName: proj, emailChan: this.emailChans[proj],
				mutex: workersMutex}
			worker.init(section)
			this.workers[proj][section.String("camel_name", "")] = worker
		}
	}

}

func (this *AlarmOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
	}

	var (
		pack       *engine.PipelinePack
		reloadChan = make(chan interface{})
		ok         = true
		inChan     = r.InChan()
	)

	// start all the workers
	for _, workers := range this.workers {
		for _, w := range workers {
			go w.run(e)
		}
	}

	for projName, emailChan := range this.emailChans {
		go this.runSendAlarmsWatchdog(e.Project(projName), emailChan)
	}

	observer.Subscribe(engine.RELOAD, reloadChan)

	for ok {
		select {
		case <-reloadChan:
			// TODO

		case pack, ok = <-inChan:
			if !ok {
				break
			}

			this.handlePack(r, e, pack)
			pack.Recycle()
		}
	}

	// start all the workers
	for _, workers := range this.workers {
		for _, w := range workers {
			w.stop()
		}
	}

	return nil
}

func (this *AlarmOutput) sendAlarmMailsLoop(project *engine.ConfProject,
	mailBody *bytes.Buffer, bodyLines *int) {
	const mailTitlePrefix = "ALS Alarm"
	var (
		mailConf          = project.MailConf
		mailSleep         = mailConf.SleepStart
		busyLineThreshold = mailConf.BusyLineThreshold
		bodyLineThreshold = mailConf.LineThreshold
		maxSleep          = mailConf.SleepMax
		minSleep          = mailConf.SleepMin
		sleepStep         = mailConf.SleepStep
	)

	for {
		select {
		case <-time.After(time.Second * time.Duration(mailSleep)):
			if *bodyLines >= bodyLineThreshold {
				go Sendmail(mailConf.Recipients,
					fmt.Sprintf("%s - %d events", mailTitlePrefix, *bodyLines),
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
		bodyLines int
		mailBody  bytes.Buffer
	)

	go this.sendAlarmMailsLoop(project, &mailBody, &bodyLines)

	for line := range emailChan {
		if globals.Debug {
			project.Printf("got email alarm: %s\n", line)
		}

		mailBody.WriteString(fmt.Sprintf("%s %s\n",
			gotime.TsToString(int(time.Now().UTC().Unix())), line))
		bodyLines += 1
	}

}

func (this *AlarmOutput) handlePack(r engine.OutputRunner, e *engine.EngineConfig,
	pack *engine.PipelinePack) {
	this.workers[pack.Project][pack.Logfile.CamelCaseName()].inject(pack.Message)
}

func init() {
	engine.RegisterPlugin("AlarmOutput", func() engine.Plugin {
		return new(AlarmOutput)
	})
}
