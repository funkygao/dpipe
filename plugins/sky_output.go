package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	sky "github.com/funkygao/skyapi"
	"strconv"
)

type SkyOutput struct {
	table    *sky.Table
	uidField string
	project  string
}

func (this *SkyOutput) Init(config *conf.Conf) {
	this.uidField = config.String("uid_field", "_log_info.uid")
	this.project = config.String("project", "")
	var (
		host string = config.String("host", "localhost")
		port int    = config.Int("port", 8585)
	)
	client := sky.NewClient(host)
	client.Port = port

	if !client.Ping() {
		panic(fmt.Sprintf("sky server not running: %s:%d", host, port))
	}

	this.table, _ = client.GetTable(config.String("table", ""))
	if this.table == nil {
		panic("must create table in advance")
	}

}

func (this *SkyOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	var (
		ok      = true
		pack    *engine.PipelinePack
		inChan  = r.InChan()
		globals = engine.Globals()
		project = h.Project(this.project)
	)

LOOP:
	for ok {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break LOOP
			}

			if globals.Debug {
				globals.Println(*pack)
			}

			this.feedSky(project, pack)
			pack.Recycle()
		}
	}

	return nil
}

func (this *SkyOutput) feedSky(project *engine.ConfProject,
	pack *engine.PipelinePack) {
	var (
		uid interface{}
		err error
	)

	// get uid
	uid, err = pack.Message.FieldValue(this.uidField, als.KEY_TYPE_INT)
	if err != nil {
		if project.ShowError {
			project.Printf("invalid uid: %v %s", err, *pack)
		}

		return
	}

	eventMap, err := pack.Message.Map()
	if err != nil {
		if project.ShowError {
			project.Println(err)
		}

		return
	}

	event := sky.NewEvent(pack.Message.Time(), eventMap)

	// objectId is uid string
	err = this.table.AddEvent(strconv.Itoa(uid.(int)), event, sky.Merge)
	if err != nil && project.ShowError {
		project.Println(err)
	}
}

func init() {
	engine.RegisterPlugin("SkyOutput", func() engine.Plugin {
		return new(SkyOutput)
	})
}
