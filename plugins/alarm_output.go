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

type alarmProjectMailConf struct {
	recipients        string
	severityPoolSize  int
	severityThreshold int
	suppressHours     []int
	interval          int
}

type alarmProjectConf struct {
	name     string
	mailConf alarmProjectMailConf
	workers  map[string]*alarmWorker // key is camelName

	emailChan chan alarmMailMessage
	stopChan  chan interface{}
}

func (this *alarmProjectConf) fromConfig(config *conf.Conf) {
	this.name = config.String("name", "")
	if this.name == "" {
		panic("project has no 'name'")
	}

	mailSection, err := config.Section("alarm_email")
	if err == nil {
		this.mailConf = alarmProjectMailConf{}
		this.mailConf.severityPoolSize = mailSection.Int("severity_pool_size", 100)
		this.mailConf.severityThreshold = mailSection.Int("severity_threshold", 8)
		this.mailConf.suppressHours = mailSection.IntList("suppress_hours", nil)
		this.mailConf.recipients = mailSection.String("recipients", "")
		if this.mailConf.recipients == "" {
			panic("mail alarm can't have no recipients")
		}
		this.mailConf.interval = mailSection.Int("interval", 300)
	}

	this.emailChan = make(chan alarmMailMessage)
	workersMutex := new(sync.Mutex)
	this.workers = make(map[string]*alarmWorker)
	for i := 0; i < len(config.List("workers", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("workers[%d]", i))
		if err != nil {
			panic(err)
		}

		worker := &alarmWorker{projName: this.name,
			emailChan: this.emailChan, workersMutex: workersMutex}
		worker.init(section, this.stopChan)
		this.workers[worker.conf.camelName] = worker
	}
	if len(this.workers) == 0 {
		panic(fmt.Sprintf("%s empty 'workers'", this.name))
	}
}

type AlarmOutput struct {
	projects map[string]alarmProjectConf // key is project name
	stopChan chan interface{}
}

func (this *AlarmOutput) Init(config *conf.Conf) {
	this.stopChan = make(chan interface{})
	this.projects = make(map[string]alarmProjectConf)
	for i := 0; i < len(config.List("projects", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("projects[%d]", i))
		if err != nil {
			panic(err)
		}

		project := alarmProjectConf{}
		project.stopChan = this.stopChan
		project.fromConfig(section)
		if _, present := this.projects[project.name]; present {
			panic("dup project: " + project.name)
		}
		this.projects[project.name] = project
	}
	if len(this.projects) == 0 {
		panic("empty projects")
	}
}

func (this *AlarmOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		pack       *engine.PipelinePack
		reloadChan = make(chan interface{})
		ok         = true
		inChan     = r.InChan()
	)

	for name, project := range this.projects {
		go this.runSendAlarmsWatchdog(h.Project(name), project)
	}

	// start all the workers
	goAhead := make(chan bool)
	for _, project := range this.projects {
		for _, worker := range project.workers {
			go worker.run(h, goAhead)
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

			this.handlePack(pack, h)
			pack.Recycle()
		}
	}

	close(this.stopChan)

	// all the workers cleanup
	for _, project := range this.projects {
		for _, worker := range project.workers {
			worker.cleanup()
		}
	}

	for _, project := range this.projects {
		close(project.emailChan)
	}

	return nil
}

func (this *AlarmOutput) handlePack(pack *engine.PipelinePack,
	h engine.PluginHelper) {
	if worker, present := this.projects[pack.Project].
		workers[pack.Logfile.CamelCaseName()]; present {
		worker.inject(pack.Message, h.Project(pack.Project))
	}
}

func (this *AlarmOutput) runSendAlarmsWatchdog(project *engine.ConfProject,
	config alarmProjectConf) {
	var (
		mailQueue   = pqueue.New()
		mailBody    bytes.Buffer
		lastSending time.Time
		mailLine    interface{}
	)

	suppressedHour := func(hour int) bool {
		// At night we are sleeping and will never checkout the alarms
		// So queue it up till we've got up from bed
		// FIXME will the mail queue overflow?
		for _, h := range config.mailConf.suppressHours {
			if hour == h {
				return true
			}
		}

		return false
	}

	heap.Init(mailQueue)

	for alarmMessage := range config.emailChan {
		if alarmMessage.severity < config.mailConf.severityThreshold {
			// ignore little severity messages
			continue
		}

		// enque
		heap.Push(mailQueue,
			&pqueue.Item{
				Value: fmt.Sprintf("%s[%3d] %s\n",
					bjtime.TimeToString(alarmMessage.receivedAt),
					alarmMessage.severity, alarmMessage.msg),
				Priority: alarmMessage.severity})

		// check if send it out now
		if !suppressedHour(bjtime.NowBj().Hour()) &&
			mailQueue.PrioritySum() >= config.mailConf.severityPoolSize {
			if !lastSending.IsZero() &&
				time.Since(lastSending).Seconds() < float64(config.mailConf.interval) {
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

			go Sendmail(config.mailConf.recipients,
				fmt.Sprintf("ALS[%s] alarms", project.Name), mailBody.String())

			project.Printf("alarm sent=> %s", config.mailConf.recipients)

			mailBody.Reset()
			lastSending = time.Now()
		}
	}
}

func init() {
	engine.RegisterPlugin("AlarmOutput", func() engine.Plugin {
		return new(AlarmOutput)
	})
}
