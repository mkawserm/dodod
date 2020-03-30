package dodod

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mkawserm/bdodb"
)

type BleveIndexOpener struct {
	EngineName string
}

func (b *BleveIndexOpener) SetEngineName(name string) {
	b.EngineName = name
}

func (b *BleveIndexOpener) BleveIndex(dbPath string,
	indexMapping *mapping.IndexMappingImpl,
	indexName string,
	config map[string]interface{}) (bleve.Index, error) {
	if b.EngineName == "" {
		b.EngineName = bdodb.EngineName
	}

	index, err := bleve.NewUsing(dbPath, indexMapping, indexName, b.EngineName, config)

	if err != nil && err == bleve.ErrorIndexPathExists {
		index, err = bleve.OpenUsing(dbPath, config)
	}

	return index, err
}
