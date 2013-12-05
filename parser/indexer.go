package parser

import (
	"crypto/rand"
	"encoding/hex"
	json "github.com/bitly/go-simplejson"
	"github.com/funkygao/alser/config"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"time"
)

type indexEntry struct {
	typ  string
	date *time.Time
	data *json.Json
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
	if !this.conf.Bool("indexer.enabled", true) {
		logger.Println("indexer disabled")
		return
	}

	api.Domain = this.conf.String("indexer.domain", "localhost")
	api.Port = this.conf.String("indexer.port", "9200")
	this.indexName = this.conf.String("indexer.index", "rs")

	done := make(chan bool)
	indexor := core.NewBulkIndexor(this.conf.Int("indexer.bulk_max_conn", 10))
	indexor.BulkMaxDocs /= 2   // default is 100, it has mem leakage, so we lower it
	indexor.BulkMaxBuffer /= 2 // default is 1MB
	indexor.Run(done)

	for item := range this.c {
		this.store(indexor, item)
	}

	indexor.Flush()
	done <- true
}

func (this *Indexer) store(indexor *core.BulkIndexor, item indexEntry) {
	docId, err := this.genUUID()
	if err != nil {
		panic(err)
	}

	if debug {
		logger.Printf("to index[%s] type=%s %v\n", this.indexName, item.typ, *item.data)
	}

	jsonData, err := item.data.MarshalJSON()
	if err != nil {
		panic(err)
	}

	err = indexor.Index(this.indexName, item.typ, docId, "", item.date, jsonData) // ttl empty
	if err != nil {
		logger.Printf("index error[%s] %s %#v\n", item.typ, err, *item.data)
	}

	if debug {
		logger.Printf("done index[%s] type=%s %v\n", this.indexName, item.typ, *item.data)
	}

}
