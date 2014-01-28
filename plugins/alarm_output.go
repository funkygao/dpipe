package plugins

import (
	"bytes"
	"container/heap"
	"fmt"
	"github.com/funkygao/dpipe/engine"
	"github.com/funkygao/golib/bjtime"
	"github.com/funkygao/golib/observer"
	"github.com/funkygao/golib/pqueue"
	conf "github.com/funkygao/jsconf"
	"sync"
	"time"
)

type alarmMailMessage struct {
	msg        string
	severity   int
	receivedAt time.Time
}

func (this alarmMailMessage) String() string {
	return fmt.Sprintf("[%d]%s", this.severity, this.msg)
}

type AlarmOutput struct {
	// {project: chan}
	emailChans map[string]chan alarmMailMessage

	// {project: {camelName: worker}}
	workers map[string]map[string]*alarmWorker
}

func (this *AlarmOutput) Init(config *conf.Conf) {
	this.emailChans = make(map[string]chan alarmMailMessage)
	this.workers = make(map[string]map[string]*alarmWorker)

	for i := 0; i < len(config.List("projects", nil)); i++ {
		keyPrefix := fmt.Sprintf("projects[%d].", i)
		proj := config.String(keyPrefix+"name", "")
		this.emailChans[proj] = make(chan alarmMailMessage, 20)
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

LOOP:
	for ok {
		select {
		case <-reloadChan:
			// TODO

		case pack, ok = <-inChan:
			if !ok {
				break LOOP
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
	queue *pqueue.PriorityQueue) {
	var (
		globals       = engine.Globals()
		mailConf      = project.MailConf
		mailSleep     = mailConf.SleepStart
		mailBody      bytes.Buffer
		bodyLinesN    int
		totalSeverity int
		mailLine      interface{}
	)

	for !globals.Stopping {
		select {
		case <-time.After(time.Second * time.Duration(mailSleep)):
			totalSeverity = queue.PrioritySum()
			if totalSeverity > project.MailConf.SeverityThreshold {
				bodyLinesN = queue.Len()

				// backoff sleep
				if bodyLinesN >= mailConf.BusyLineThreshold {
					mailSleep -= mailConf.SleepStep
					if mailSleep < mailConf.SleepMin {
						mailSleep = mailConf.SleepMin
					}
				} else {
					// idle alarm
					mailSleep += mailConf.SleepStep
					if mailSleep > mailConf.SleepMax {
						mailSleep = mailConf.SleepMax
					}
				}

				// gather mail body content
				for {
					if queue.Len() == 0 {
						break
					}

					mailLine = heap.Pop(queue)
					mailBody.WriteString(mailLine.(*pqueue.Item).Value.(string))
				}

				go Sendmail(mailConf.Recipients,
					fmt.Sprintf("ALS[%s] - %d alarms(within %s), severity=%d",
						project.Name, bodyLinesN,
						time.Duration(mailSleep)*time.Second, totalSeverity),
					mailBody.String())
				project.Printf("alarm sent=> %s, sleep=%d\n",
					mailConf.Recipients, mailSleep)

				mailBody.Reset()
			}
		}
	}
}

func (this *AlarmOutput) runSendAlarmsWatchdog(project *engine.ConfProject,
	emailChan chan alarmMailMessage) {
	var (
		mailQueue = pqueue.New()
	)

	heap.Init(mailQueue)

	go this.sendAlarmMailsLoop(project, mailQueue)

	for alarmMessage := range emailChan {
		heap.Push(mailQueue,
			&pqueue.Item{
				Value: fmt.Sprintf("%s[%3d] %s\n",
					bjtime.TimeToString(alarmMessage.receivedAt),
					alarmMessage.severity, alarmMessage.msg),
				Priority: alarmMessage.severity})
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
