package engine

import (
	conf "github.com/funkygao/jsconf"
	"log"
	"os"
)

type projectEmailConf struct {
	Recipients                                string
	BusyLineThreshold                         int
	SleepStart, SleepStep, SleepMax, SleepMin int
	LineThreshold                             int
}

type ConfProject struct {
	*log.Logger

	Name        string `json:"name"`
	IndexPrefix string `json:"index_prefix"`
	ShowError   bool

	MailConf projectEmailConf
}

func (this *ConfProject) FromConfig(c *conf.Conf) {
	this.Name = c.String("name", "")
	this.IndexPrefix = c.String("index_prefix", this.Name)
	this.ShowError = c.Bool("show_error", true)
	mailSection, err := c.Section("alarm_email")
	if err == nil {
		this.MailConf = projectEmailConf{}
		this.MailConf.Recipients = mailSection.String("recipients", "")
		this.MailConf.LineThreshold = mailSection.Int("line_threshold", 10)
		this.MailConf.BusyLineThreshold = mailSection.Int("busy_line_threshold", 25)
		this.MailConf.SleepStart = mailSection.Int("sleep_start", 600)
		this.MailConf.SleepMin = mailSection.Int("sleep_min", 240)
		this.MailConf.SleepMax = mailSection.Int("sleep_max", 1600)
		this.MailConf.SleepStep = mailSection.Int("sleep_step", 60)
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

func (this *ConfProject) Stop() {
	this.Println("Stopped")
}
