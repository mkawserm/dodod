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

type IdMatch struct {
	Id    string  `json:"id"`
	Score float64 `json:"score"`
}

type FindId interface {
	FindId(queryInput string, offset int) (
		total uint64,
		queryTime time.Duration,
		result []*IdMatch,
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

type FaceInput struct {
	FacetName  string `json:"facetName"`
	QueryInput string `json:"queryInput"`
	FacetLimit int    `json:"facetLimit"`
}

type FacetOutput struct {
	TermName  string `json:"termName"`
	TermCount int    `json:"termCount"`
}

type FacetSearch interface {
	FacetSearch(facetInput []FaceInput) (
		queryTime time.Duration,
		data map[string][]FacetOutput,
		err error)
}
