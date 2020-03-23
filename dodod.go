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

type DbCredentials interface {
	ReadPath() (dbPath string, err error)
	ReadPassword() (password string, err error)
}

type ModelRegistry interface {
	RegisterModel(model interface{}) error
}

type Query interface {
	Read(data []interface{}) (uint64, error)
}

type Mutation interface {
	Create(data []interface{}) error
	Update(data []interface{}) error
	Delete(data []interface{}) error
}

type Search interface {
	Search(query string, from uint64) (total uint64, result []interface{}, err error)
}

type ComplexSearch interface {
	ComplexSearch(query, sortBy, queryType string,
		from uint64, limit uint64) (total uint64, result []interface{}, err error)
}

type Dodod interface {
	Query
	Mutation
	Search
	ComplexSearch
	ModelRegistry
}
