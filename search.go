package dodod

import (
	"context"
	"time"
)
import "github.com/blevesearch/bleve"

type Search interface {
	Search(queryInput string, offset int) (
		total uint64,
		queryTime time.Duration,
		result []interface{},
		err error)
}

type ComplexSearch interface {
	ComplexSearch(
		queryInput string,
		sortBy []string,
		queryType string,
		offset int, limit int) (
		total uint64,
		queryTime time.Duration,
		result []interface{},
		err error)
}

type BleveSearch interface {
	BleveSearch(req *bleve.SearchRequest) (*bleve.SearchResult, error)

	BleveSearchInContext(
		ctx context.Context,
		req *bleve.SearchRequest) (*bleve.SearchResult, error)
}
