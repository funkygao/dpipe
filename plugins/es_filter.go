package plugins

import (
	"fmt"
	"github.com/funkygao/als"
	"github.com/funkygao/dpipe/engine"
	conf "github.com/funkygao/jsconf"
	"time"
)

type esConverter struct {
	keys        []string // key name
	typ         string   // type
	currency    string   // currency field name
	rang        []int    // range
	normalizers []string
}

func (this *esConverter) load(section *conf.Conf) {
	this.keys = section.StringList("keys", nil)
	this.typ = section.String("type", "")
	this.currency = section.String("currency", "")
	this.rang = section.IntList("range", nil)
	this.normalizers = section.StringList("normalizers", nil)
}

type EsFilter struct {
	ident        string
	indexPattern string
	converters   []esConverter
}

func (this *EsFilter) Init(config *conf.Conf) {
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
	this.converters = make([]esConverter, 0, 10)
	this.indexPattern = config.String("index_pattern", "")
	for i := 0; i < len(config.List("converts", nil)); i++ {
		section, err := config.Section(fmt.Sprintf("%s[%d]", "converts", i))
		if err != nil {
			panic(err)
		}

		c := esConverter{}
		c.load(section)
		this.converters = append(this.converters, c)
	}

	geodbFile := config.String("geodbfile", "")
	if err := als.LoadGeoDb(geodbFile); err != nil {
		panic(err)
	}
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("Loaded geodb %s\n", geodbFile)
	}
}

func (this *EsFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	var (
		globals = engine.Globals()
		pack    *engine.PipelinePack
		ok      = true
		count   = 0
		inChan  = r.InChan()
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

			if this.handlePack(pack, h.Project(pack.Project)) {
				count += 1
				r.Inject(pack)
			} else {
				pack.Recycle()
			}
		}
	}

	globals.Printf("[%s]Total filtered: %d", r.Name(), count)

	return nil
}

func (this *EsFilter) handlePack(pack *engine.PipelinePack, project *engine.ConfProject) bool {
	if pack.EsType == "" {
		pack.EsType = pack.Logfile.CamelCaseName()
	}
	if pack.EsIndex == "" {
		pack.EsIndex = indexName(project, this.indexPattern,
			time.Unix(int64(pack.Message.Timestamp), 0))
	}

	// each ES item has area and ts fields
	pack.Ident = this.ident
	pack.Message.SetField("_area", pack.Message.Area)
	pack.Message.SetField("_t", pack.Message.Timestamp)

	for _, conv := range this.converters {
		for _, key := range conv.keys {
			if conv.normalizers != nil {
				for _, norm := range conv.normalizers {
					val, err := pack.Message.FieldValue(key, als.KEY_TYPE_STRING)
					if err != nil {
						// no such field
						break
					}

					normed := normalizers[norm].ReplaceAll([]byte(val.(string)),
						[]byte("?"))
					val = string(normed)
					pack.Message.SetField(key+"_norm", val)
				}

				continue
			}

			switch conv.typ {
			case "money":
				amount, err := pack.Message.FieldValue(key, als.KEY_TYPE_MONEY)
				if err != nil {
					// has no such field
					continue
				}

				currency, err := pack.Message.FieldValue(conv.currency, als.KEY_TYPE_STRING)
				if err != nil {
					// has money field, but no currency field?
					return false
				}

				pack.Message.SetField("_usd",
					als.MoneyInUsdCents(currency.(string), amount.(int)))

			case "ip":
				ip, err := pack.Message.FieldValue(key, als.KEY_TYPE_IP)
				if err != nil {
					continue
				}

				pack.Message.SetField("_cntry", als.IpToCountry(ip.(string)))

			case "range":
				if len(conv.rang) < 2 {
					continue
				}

				val, err := pack.Message.FieldValue(key, als.KEY_TYPE_INT)
				if err != nil {
					continue
				}

				pack.Message.SetField(key+"_rg", als.GroupInt(val.(int), conv.rang))

			case "del":
				pack.Message.DelField(key)
			}
		}

	}

	return true
}

func init() {
	engine.RegisterPlugin("EsFilter", func() engine.Plugin {
		return new(EsFilter)
	})
}
