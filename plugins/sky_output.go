package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	sky "github.com/funkygao/skyapi"
	"strconv"
)

type skyOutputField struct {
	camelName string
	action    string
	name      string
	typ       string
}

func (this *skyOutputField) load(section *conf.Conf) {
	this.name = section.String("name", "")
	this.typ = section.String("type", als.KEY_TYPE_STRING)
	this.camelName = section.String("camel_name", "")
	this.action = section.String("action", "")
}

type SkyOutput struct {
	table    *sky.Table
	stopChan chan bool
	fields   []skyOutputField
}

func (this *SkyOutput) Init(config *conf.Conf) {
	this.stopChan = make(chan bool)
	var (
		host string = config.String("host", "localhost")
		port int    = config.Int("port", 8585)
	)
	client := sky.NewClient(host)
	client.Port = port

	if !client.Ping() {
		panic(fmt.Sprintf("sky server not running: %s:%d", host, port))
	}

	this.table, _ = client.GetTable(config.String("table", "user"))
	if this.table == nil {
		panic("must create table in advance")
	}

	this.fields = make([]skyOutputField, 0, 10)
	for i := 0; i < len(config.List("fields", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("fields[%d]", i))
		if err != nil {
			panic(err)
		}

		f := skyOutputField{}
		f.load(section)
		this.fields = append(this.fields, f)
	}

}

func (this *SkyOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		ok     = true
		pack   *engine.PipelinePack
		inChan = r.InChan()
	)

	for ok {
		select {
		case <-this.stopChan:
			ok = false

		case pack, ok = <-inChan:
			if !ok {
				break
			}

			this.feedSky(pack)
			pack.Recycle()
		}
	}

	return nil
}

func (this *SkyOutput) feedSky(pack *engine.PipelinePack) {
	var (
		uid interface{}
		val interface{}
		err error
	)

	// get uid
	uid, err = pack.Message.FieldValue("_log_info.uid", als.KEY_TYPE_INT)
	if err != nil {
		uid, err = pack.Message.FieldValue("uid", als.KEY_TYPE_INT)
		if err != nil {
			// if not uid based, ignored
			return
		}

	}

	event := sky.NewEvent(pack.Message.Time(), map[string]interface{}{})
	// fill in the event fields
	for _, f := range this.fields {
		if pack.Logfile.CamelCaseName() != f.camelName {
			continue
		}

		if f.action != "" {
			// already specified dedicated action
			event.Data["action"] = f.action
			event.Data["area"] = pack.Message.Area
		} else {
			val, err = pack.Message.FieldValue(f.name, f.typ)
			if err != nil {
				continue
			}
		}

		event.Data[f.name] = val
	}

	// objectId is uid string
	this.table.AddEvent(strconv.Itoa(uid.(int)), event, sky.Merge)
}

func init() {
	engine.RegisterPlugin("SkyOutput", func() engine.Plugin {
		return new(SkyOutput)
	})
}
