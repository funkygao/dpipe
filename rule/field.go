package rule

import (
	json "github.com/bitly/go-simplejson"
	"path/filepath"
	"regexp"
	"strings"
)

// Key data sink to 4 kinds of targets
// ======== ========== ==============
// sqldb(d) indexer(i) sink
// ======== ========== ==============
//        Y Y          3, default
//        Y N          2, alarm only
//        N Y          1, index only
//        N N          0, validator only
// ======== ========== ==============
type Field struct {
	Name    string
	Type    string // float, string(default), int, money
	Contain string // only being validator instead of data
	Ignores []string
	Filters []string // currently not used yet TODO
	Regex   []string
	Sink    int // bit op
}

func (this *Field) MsgIgnored(msg string) bool {
	for _, ignore := range this.Ignores {
		if strings.Contains(msg, ignore) {
			return true
		}

		if strings.HasPrefix(ignore, "regex:") {
			pattern := strings.TrimSpace(ignore[6:])
			// TODO lessen the overhead
			if matched, err := regexp.MatchString(pattern, msg); err == nil && matched {
				return true
			}
		}
	}

	// filters means only when the key satisfy at least one of the filter rule
	// will the msg be accepted
	if this.Filters != nil {
		for _, f := range this.Filters {
			if msg == f {
				return false
			}
		}

		return true
	}

	return false
}

func (this *Field) Alarmable() bool {
	return this.Sink&2 != 0
}

func (this *Field) Indexable() bool {
	return this.Sink&1 != 0
}

func (this *Field) IsIP() bool {
	return this.Name == FIELD_NAME_IP || this.Type == FIELD_TYPE_IP
}

func (this *Field) IsMoney() bool {
	return this.Type == FIELD_TYPE_MONEY
}

func (this *Field) IsFloat() bool {
	return this.Type == FIELD_TYPE_FLOAT
}

func (this *Field) IsInt() bool {
	return this.Type == FIELD_TYPE_INT
}

func (this *Field) IsLevel() bool {
	return this.Type == FIELD_TYPE_LEVEL
}

func (this *Field) IsString() bool {
	return this.Type == FIELD_TYPE_STRING
}

func (this *Field) IsBaseFile() bool {
	return this.Type == FIELD_TYPE_BASEFILE
}

func (this *Field) JsonValue(data *json.Json) (val interface{}, err error) {
	switch {
	case this.IsString(), this.IsIP():
		val, err = data.Get(this.Name).String()
	case this.IsFloat():
		val, err = data.Get(this.Name).Float64()
	case this.IsInt(), this.IsMoney(), this.IsLevel():
		val, err = data.Get(this.Name).Int()
	case this.IsBaseFile():
		var fullpath string
		fullpath, err = data.Get(this.Name).String()
		if err != nil {
			return
		}
		val = filepath.Base(fullpath)
	default:
		panic("invalid field type: " + this.Type)
	}

	return
}
