package plugins

import (
	"fmt"
	"github.com/funkygao/funpipe/engine"
	conf "github.com/funkygao/jsconf"
)

type cardinalityField struct {
	key      string
	interval []string
}

type cardinalityConverter struct {
	logPrefix string
	project   string // action
	fields    []cardinalityField
}

func (this *cardinalityConverter) load(section *conf.Conf) {
	this.logPrefix = section.String("log_prefix", "")
	this.project = section.String("proj", "")
	this.fields = make([]cardinalityField, 0, 5)
	for i := 0; i < len(section.List("fields", nil)); i++ {
		keyPrefix := fmt.Sprintf("fields[%d].", i)
		field := cardinalityField{}
		field.key = section.String(keyPrefix+"key", "")
		field.interval = section.StringList(keyPrefix+"interval", nil)
		this.fields = append(this.fields, field)
	}
}

type CardinalityFilter struct {
	sink       string
	converters []cardinalityConverter
}

func (this *CardinalityFilter) Init(config *conf.Conf) {
	const CONV = "converts"
	this.sink = config.String("sink", "")
	for i := 0; i < len(config.List(CONV, nil)); i++ {
		section, err := config.Section(fmt.Sprintf("%s[%d]", CONV, i))
		if err != nil {
			panic(err)
		}

		c := cardinalityConverter{}
		c.load(section)
		this.converters = append(this.converters, c)
	}
}

func (this *CardinalityFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	globals := engine.Globals()
	if globals.Verbose {
		globals.Printf("[%s] started\n", r.Name())
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
		}
	}

	return nil
}

// for each inbound pack, this filter will generate several new pack
// the original pack will be recycled immediately
func (this *CardinalityFilter) handlePack(r engine.FilterRunner,
	h engine.PluginHelper, pack *engine.PipelinePack) {
	globals := engine.Globals()
	for _, c := range this.converters {
		if !pack.Logfile.MatchPrefix(c.logPrefix) || pack.Project != c.project {
			continue
		}

		for _, f := range c.fields {
			val, err := pack.Message.ValueOfKey(f.key)
			if err != nil {
				if globals.Verbose {
					globals.Println(err)
				}

				return
			}

			for _, interval := range f.interval {
				// generate new pack
				p := h.PipelinePack(0)
				p.CardinalityKey = fmt.Sprintf("%s.%s.%s", pack.Project, f.key, interval)
				p.CardinalityData = val
				p.CardinalityInterval = interval

				r.Inject(p)
			}
		}
	}

	pack.Recycle()
}

func init() {
	engine.RegisterPlugin("CardinalityFilter", func() engine.Plugin {
		return new(CardinalityFilter)
	})
}
