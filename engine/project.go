package engine

import (
	"fmt"
	conf "github.com/funkygao/jsconf"
	"log"
	"os"
)

type ConfProject struct {
	*log.Logger

	Name        string
	IndexPrefix string
}

func (this *ConfProject) FromConfig(c *conf.Conf) {
	this.Name = c.String("name", "")
	this.IndexPrefix = c.String("index_prefix", this.Name)

	logfile := c.String("logfile", "var/"+this.Name+".log")
	logWriter, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	logOptions := log.Ldate | log.Ltime
	if Globals().Debug {
		logOptions |= log.Lshortfile | log.Lmicroseconds
	}

	this.Logger = log.New(logWriter, fmt.Sprintf("[%d]", os.Getpid()), logOptions)
}

func (this *ConfProject) Stop() {

}
