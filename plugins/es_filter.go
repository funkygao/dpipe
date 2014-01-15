package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"time"
)

type esConverter struct {
	key    string // key name
	action string // action
	cur    string // currency
	rang   []int  // range
}

func (this *esConverter) load(section *conf.Conf) {
	this.key = section.String("key", "")
	this.action = section.String("act", "")
	this.cur = section.String("cur", "")
	this.rang = section.IntList("range", nil)
}

type EsFilter struct {
	ident        string
	indexPattern string
	converters   []esConverter
}

func (this *EsFilter) Init(config *conf.Conf) {
	const CONV = "converts"
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
	this.converters = make([]esConverter, 0, 10)
	this.indexPattern = config.String("index_pattern", "")
	for i := 0; i < len(config.List(CONV, nil)); i++ {
		section, err := config.Section(fmt.Sprintf("%s[%d]", CONV, i))
		if err != nil {
			panic(err)
		}

		c := esConverter{}
		c.load(section)
		this.converters = append(this.converters, c)
	}
}

func (this *EsFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	globals := engine.Globals()
	geodbFile := h.EngineConfig().String("geodbfile", "")
	if err := als.LoadGeoDb(geodbFile); err != nil {
		panic(err)
	}
	if globals.Verbose {
		globals.Printf("Load geodb from %s\n", geodbFile)
	}

	var (
		pack   *engine.PipelinePack
		ok     = true
		inChan = r.InChan()
	)

	for ok && !globals.Stopping {
		select {
		case pack, ok = <-inChan:
			if !ok {
				break
			}

			this.handlePack(r, h, pack)
			pack.Recycle()
		}
	}

	return nil
}

func (this *EsFilter) handlePack(r engine.FilterRunner,
	h engine.PluginHelper, pack *engine.PipelinePack) {
	project := h.Project(pack.Project)

	if pack.EsType == "" {
		pack.EsType = pack.Logfile.CamelCaseName()
	}
	if pack.EsIndex == "" {
		pack.EsIndex = indexName(project, this.indexPattern,
			time.Unix(int64(pack.Message.Timestamp), 0))
	}

	if pack.EsType == "" {
		engine.Globals().Printf("%s %v\n", pack.EsType, *pack)
		return
	}
	if pack.EsIndex == "" {
		engine.Globals().Printf("%s %v\n", pack.EsIndex, *pack)
		return
	}

	p := h.PipelinePack(pack.MsgLoopCount)
	p.Ident = this.ident
	p.EsIndex = pack.EsIndex
	p.Project = pack.Project
	p.EsType = pack.EsType

	// each ES item has area and ts fields
	p.Message.SetField("area", pack.Message.Area)
	p.Message.SetField("ts", pack.Message.Timestamp)

	for _, conv := range this.converters {
		switch conv.action {
		case "money":
			amount, err := pack.Message.FieldValue(conv.key, als.KEY_TYPE_MONEY)
			if err != nil {
				// has no such field
				continue
			}

			currency, err := pack.Message.FieldValue(conv.cur, als.KEY_TYPE_STRING)
			if err != nil {
				// has money field, but no currency field?
				return
			}

			p.Message.SetField("usd",
				als.MoneyInUsdCents(currency.(string), amount.(int)))

		case "ip":
			ip, err := pack.Message.FieldValue(conv.key, als.KEY_TYPE_IP)
			if err != nil {
				continue
			}

			p.Message.SetField("cntry", als.IpToCountry(ip.(string)))

		case "range":
			if len(conv.rang) < 2 {
				continue
			}

			val, err := pack.Message.FieldValue(conv.key, als.KEY_TYPE_INT)
			if err != nil {
				continue
			}

			p.Message.SetField(conv.key+"_rg", als.GroupInt(val.(int), conv.rang))

		case "del":
			pack.Message.DelField(conv.key)
		}
	}

	r.Inject(p)
}

func init() {
	engine.RegisterPlugin("EsFilter", func() engine.Plugin {
		return new(EsFilter)
	})
}
