package parser

import (
	"crypto/rand"
	"encoding/hex"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
)

type indexEntry struct {
	typ  string
	data json.Json
}

type Indexer struct {
	c         chan indexEntry
	indexName string // index name
	conf      *config.Config
}

func newIndexer(conf *config.Config) (this *Indexer) {
	this = new(Indexer)
	this.conf = conf
	this.c = make(chan indexEntry, 1000)

	return
}

// 1914 ns/op from BenchmarkUUID
func (this *Indexer) genUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}

	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil
}

func (this *Indexer) mainLoop() {
	api.Domain = this.conf.String("indexer.domain", "localhost")
	api.Port = this.conf.String("indexer.port", "9200")
	this.indexName = this.conf.String("indexer.index", "rs")

	for item := range this.c {
		this.store(item)
	}
}

func (this *Indexer) store(item indexEntry) {
	id, err := this.genUUID()
	if err != nil {
		panic(err)
	}

	if debug {
		logger.Printf("index[%s] type=%s %v\n", this.indexName, item.typ, item.data)
	}

	response, err := core.Index(false, this.indexName, item.typ, id, item.data)
	if err != nil || !response.Ok {
		logger.Printf("index error[%s] %s %#v\n", item.typ, err, response)
	}
}

func (this *Indexer) index(item indexEntry) {
	this.c <- item
}
