/*
flashlog is saved in mysql
+-------------+---------------------+------+-----+---------+----------------+
| Field       | Type                | Null | Key | Default | Extra          |
+-------------+---------------------+------+-----+---------+----------------+
| id          | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
| uid         | bigint(20) unsigned | NO   | MUL | NULL    |                |
| type        | int(10) unsigned    | NO   | MUL | NULL    |                |
| data        | blob                | NO   |     | NULL    |                |
| ip          | bigint(20)          | NO   | MUL | NULL    |                |
| ua          | int(10) unsigned    | NO   | MUL | NULL    |                |
| date_create | int(10) unsigned    | NO   | MUL | NULL    |                |
+-------------+---------------------+------+-----+---------+----------------+
*/
package plugins

import (
	"github.com/funkygao/funpipe/engine"
)

type FlashlogInputConfig struct {
	dsn string
}

type FlashlogInput struct {
	*FlashlogInputConfig
}

func (this *FlashlogInput) Init(config interface{}) {
	c := config.(*FlashlogInputConfig)
	this.FlashlogInputConfig = c

}

func (this *FlashlogInput) Config() interface{} {
	return FlashlogInputConfig{
		dsn: "flashlog:flashlog@unix(/var/run/mysqld/mysqld.sock)/flashlog?charset=utf8",
	}
}

func (this *FlashlogInput) Run(r engine.InputRunner, e *engine.EngineConfig) error {
	return nil
}

func init() {
	engine.RegisterPlugin("FlashlogInput", func() interface{} {
		return new(FlashlogInput)
	})
}
