package engine

import (
	conf "github.com/funkygao/jsconf"
	"log"
	"os"
)

type projectEmailConf struct {
	Recipients        string
	SeverityPoolSize  int
	SeverityThreshold int
	SuppressHours     []int
	Interval          int
}

type ConfProject struct {
	*log.Logger

	Name        string `json:"name"`
	IndexPrefix string `json:"index_prefix"`
	ShowError   bool   `json:"show_error"`

	MailConf projectEmailConf
}

func (this *ConfProject) FromConfig(c *conf.Conf) {
	this.Name = c.String("name", "")
	this.IndexPrefix = c.String("index_prefix", this.Name)
	this.ShowError = c.Bool("show_error", true)
	mailSection, err := c.Section("alarm_email")
	if err == nil {
		this.MailConf = projectEmailConf{}
		this.MailConf.SeverityPoolSize = mailSection.Int("severity_pool_size", 100)
		this.MailConf.SeverityThreshold = mailSection.Int("severity_threshold", 8)
		this.MailConf.SuppressHours = mailSection.IntList("suppress_hours", nil)
		this.MailConf.Recipients = mailSection.String("recipients", "")
		this.MailConf.Interval = mailSection.Int("interval", 300)
	}

	logfile := c.String("logfile", "var/"+this.Name+".log")
	logWriter, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	logOptions := log.Ldate | log.Ltime
	if Globals().Verbose {
		logOptions |= log.Lshortfile
	}
	if Globals().Debug {
		logOptions |= log.Lmicroseconds
	}

	this.Logger = log.New(logWriter, "", logOptions)
}

func (this *ConfProject) Start() {
	this.Println("Started")
}

func (this *ConfProject) Stop() {
	this.Println("Stopped")
}
