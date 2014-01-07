package engine

import (
	"fmt"
	conf "github.com/funkygao/jsconf"
	"log"
	"os"
)

type ConfProject struct {
	*log.Logger

	Name        string `json:"name"`
	IndexPrefix string `json:"index_prefix"`

	// such as a@a.com,b@b.com
	AlarmMailTo string `json:"alarm_to"`
}

func (this *ConfProject) FromConfig(c *conf.Conf) {
	this.Name = c.String("name", "")
	this.IndexPrefix = c.String("index_prefix", this.Name)
	this.AlarmMailTo = c.String("alarm_to", "")

	logfile := c.String("logfile", "var/"+this.Name+".log")
	logWriter, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	logOptions := log.Ldate | log.Ltime | log.Lshortfile
	if Globals().Debug {
		logOptions |= log.Lmicroseconds
	}

	this.Logger = log.New(logWriter, fmt.Sprintf("[%d]", os.Getpid()), logOptions)
}

func (this *ConfProject) Stop() {

}
