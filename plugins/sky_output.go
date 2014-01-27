package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	sky "github.com/funkygao/skyapi"
	"strconv"
	"strings"
)

type SkyOutput struct {
	table           *sky.Table
	uidField        string
	uidFieldType    string
	actionField     string
	actionFieldType string
	project         string
}

func (this *SkyOutput) Init(config *conf.Conf) {
	const TYPE_SEP = ":"
	this.uidFieldType, this.actionFieldType = als.KEY_TYPE_INT, als.KEY_TYPE_STRING
	this.uidField = config.String("uid_field", "_log_info.uid")
	if strings.Contains(this.uidField, TYPE_SEP) {
		p := strings.SplitN(this.uidField, TYPE_SEP, 2)
		this.uidField, this.uidFieldType = p[0], p[1]
	}
	this.actionField = config.String("action_field", "action")
	if this.actionField == "" {
		panic("empty action field")
	}
	if strings.Contains(this.actionField, TYPE_SEP) {
		p := strings.SplitN(this.actionField, TYPE_SEP, 2)
		this.actionField, this.actionFieldType = p[0], p[1]
	}

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
		uid    interface{}
		action interface{}
		err    error
	)

	// get uid
	uid, err = pack.Message.FieldValue(this.uidField, this.uidFieldType)
	if err != nil {
		if project.ShowError {
			project.Printf("invalid uid: %v %s", err, *pack)
		}

		return
	}

	action, err = pack.Message.FieldValue(this.actionField, this.actionFieldType)
	if err != nil {
		if project.ShowError {
			project.Printf("invalid action: %v %s", err, *pack)
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
	if this.actionFieldType == als.KEY_TYPE_INT {
		event.Data["action"] = "action_" + strconv.Itoa(action.(int))
	} else {
		event.Data["action"] = action
	}

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
