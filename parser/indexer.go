package parser

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/funkygao/alser/config"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"strings"
	//"time"
)

type Indexer struct {
	lineIn    chan string // typ:line
	indexName string      // index name
	conf      *config.Config
}

func newIndexer(conf *config.Config) (this *Indexer) {
	this = new(Indexer)
	this.conf = conf
	this.lineIn = make(chan string, 1000)

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

	for line := range this.lineIn {
		this.store(line)
	}
}

func (this *Indexer) store(line string) {
	id, err := this.genUUID()
	if err != nil {
		panic(err)
	}

	parts := strings.SplitN(line, ":", 2)
	typ, data := parts[0], parts[1]
	//now := time.Now()
	logger.Println(data)
	response, err := core.Index(false, this.indexName, typ, id, data)
	if err != nil {
		panic(err)
	}

	logger.Println(response)
}

func (this *Indexer) index(typ, line string) {
	this.lineIn <- typ + ":" + line
}
