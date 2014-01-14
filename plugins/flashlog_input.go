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
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type FlashlogInput struct {
	dsn   string
	ident string
}

func (this *FlashlogInput) Init(config *conf.Conf) {
	this.dsn = config.String("dsn",
		"flashlog:flashlog@unix(/var/run/mysqld/mysqld.sock)/flashlog?charset=utf8")
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
}

func (this *FlashlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {

	return nil
}

func (this *FlashlogInput) Stop() {

}

func init() {
	engine.RegisterPlugin("FlashlogInput", func() engine.Plugin {
		return new(FlashlogInput)
	})
}
