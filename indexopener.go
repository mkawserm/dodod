package dodod

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type IndexOpener interface {
	BleveIndex(dbPath string,
		indexMapping *mapping.IndexMappingImpl,
		indexName string,
		config map[string]interface{}) (bleve.Index, error)
}
