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
		this.emailChans[proj] = make(chan alarmMailMessage, 50)
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

func (this *AlarmOutput) runSendAlarmsWatchdog(project *engine.ConfProject,
	emailChan chan alarmMailMessage) {
	var (
		mailQueue     = pqueue.New()
		totalSeverity int
		mailBody      bytes.Buffer
		lastSending   time.Time
		mailLine      interface{}
	)

	heap.Init(mailQueue)

	for alarmMessage := range emailChan {
		heap.Push(mailQueue,
			&pqueue.Item{
				Value: fmt.Sprintf("%s[%3d] %s\n",
					bjtime.TimeToString(alarmMessage.receivedAt),
					alarmMessage.severity, alarmMessage.msg),
				Priority: alarmMessage.severity})

		// check if send it out now
		totalSeverity = mailQueue.PrioritySum()
		if totalSeverity >= project.MailConf.SeverityThreshold {
			if !lastSending.IsZero() &&
				time.Since(lastSending).Seconds() < float64(project.MailConf.Interval) {
				// we can't send too many emails in emergancy
				continue
			}

			// gather mail body content
			for {
				if mailQueue.Len() == 0 {
					break
				}

				mailLine = heap.Pop(mailQueue)
				mailBody.WriteString(mailLine.(*pqueue.Item).Value.(string))
			}

			go Sendmail(project.MailConf.Recipients,
				fmt.Sprintf("ALS[%s]total severity=%d",
					project.Name, totalSeverity), mailBody.String())

			project.Printf("alarm sent=> %s", project.MailConf.Recipients)

			mailBody.Reset()
			lastSending = time.Now()
		}
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
