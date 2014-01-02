package engine

import (
	conf "github.com/funkygao/jsconf"
)

type ConfProject struct {
	Name        string
	Logger      string
	IndexPrefix string
}

func (this *ConfProject) FromConfig(c *conf.Conf) {
	this.Name = c.String("name", "")
	this.Logger = c.String("logger", "var/"+this.Name)
	this.IndexPrefix = c.String("index_prefix", this.Name)
}

func (this *ConfProject) Stop() {

}
