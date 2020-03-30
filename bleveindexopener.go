package dodod

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mkawserm/bdodb"
)

type BleveIndexOpener struct {
}

func (b *BleveIndexOpener) BleveIndex(dbPath string,
	indexMapping *mapping.IndexMappingImpl,
	indexName string,
	config map[string]interface{}) (bleve.Index, error) {

	return bdodb.BleveIndex(dbPath, indexMapping, indexName, config)
}
