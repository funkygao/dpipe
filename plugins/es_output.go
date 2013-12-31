// ElasticSerach plugin
package plugins

import (
	"errors"
	"github.com/funkygao/funpipe/engine"
)

type EsOutputConfig struct {
	indexName string
}

type EsOutput struct {
}

func (this *EsOutput) Init(config interface{}) {

}

func (this *EsOutput) Config() interface{} {

}

func (this *EsOutput) Run(r engine.OutputRunner, e *engine.EngineConfig) error {

}
