package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"strings"
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
	sink         string
	indexPattern string
	converters   []esConverter
}

func (this *EsFilter) Init(config *conf.Conf) {
	const CONV = "converts"
	this.sink = config.String("sink", "")
	if this.sink == "" {
		panic("empty sink")
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

			if this.handlePack(pack, h.Project(pack.Project)) {
				r.Inject(pack)
			} else {
				pack.Recycle()
			}
		}
	}

	return nil
}

func (this *EsFilter) indexName(project *engine.ConfProject, date time.Time) string {
	const (
		YM           = "@ym"
		INDEX_PREFIX = "fun_"
	)

	if strings.HasSuffix(this.indexPattern, YM) {
		prefix := project.IndexPrefix
		fields := strings.SplitN(this.indexPattern, YM, 2)
		if fields[0] != "" {
			// e,g. rs@ym
			prefix = fields[0]
		}

		return fmt.Sprintf("%s%s_%d_%02d", INDEX_PREFIX, prefix, date.Year(), int(date.Month()))
	}

	return INDEX_PREFIX + this.indexPattern
}

func (this *EsFilter) handlePack(pack *engine.PipelinePack, project *engine.ConfProject) bool {
	pack.Sink = this.sink
	if pack.EsType == "" {
		pack.EsType = pack.Logfile.CamelCaseName()
	}
	pack.EsIndex = this.indexName(project,
		time.Unix(int64(pack.Message.Timestamp), 0))
	if pack.EsType == "" {
		engine.Globals().Printf("%s %v\n", pack.EsType, *pack)
		return false
	}
	if pack.EsIndex == "" {
		engine.Globals().Printf("%s %v\n", pack.EsIndex, *pack)
		return false
	}

	pack.Message.SetField("area", pack.Message.Area)
	pack.Message.SetField("ts", pack.Message.Timestamp)

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
				return false
			}

			pack.Message.SetField("usd",
				als.MoneyInUsdCents(currency.(string), amount.(int)))

		case "ip":
			ip, err := pack.Message.FieldValue(conv.key, als.KEY_TYPE_IP)
			if err != nil {
				continue
			}

			pack.Message.SetField("cntry", als.IpToCountry(ip.(string)))

		case "range":
			if len(conv.rang) < 2 {
				continue
			}

			val, err := pack.Message.FieldValue(conv.key, als.KEY_TYPE_INT)
			if err != nil {
				continue
			}

			pack.Message.SetField(conv.key+"_rg", als.GroupInt(val.(int), conv.rang))

		case "del":
			pack.Message.DelField(conv.key)
		}
	}

	return true
}

func init() {
	engine.RegisterPlugin("EsFilter", func() engine.Plugin {
		return new(EsFilter)
	})
}
