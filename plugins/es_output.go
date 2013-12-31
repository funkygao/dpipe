// ElasticSerach plugin
package plugins

import (
	"errors"
	"github.com/funkygao/funpipe/engine"
)

type EsOutput struct {
	indexName, typeName string
}

func (this *EsOutput) Init(config interface{}) error {

}

func (this *EsOutput) Run(r engine.OutputRunner, c *engine.PipelineConfig) error {

}
