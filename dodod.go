package dodod

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type IndexOpener interface {
	BleveIndex(dbPath string,
		indexMappingImpl *mapping.IndexMappingImpl,
		indexName string,
		config map[string]interface{}) (bleve.Index, error)
}

type Mutation interface {
	Create(data []interface{}) error
	Update(data []interface{}) error
	Delete(data []interface{}) error
}

type Query interface {
	Read(data []interface{}) (int, error)
}

type Search interface {
	Search(q string, limit int) PageIterator
}

type PageIterator interface {
	NextPage()
	HasNextPage() bool

	TotalPage() uint64
	TotalResults() uint64

	CurrentPageNumber() uint64
	CurrentPageResults() []interface{}
}

type DbCredentials interface {
	ReadPath() (dbPath string, err error)
	ReadPassword() (password string, err error)
}
