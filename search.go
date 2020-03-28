package dodod

import (
	"context"
)

import "github.com/blevesearch/bleve"

// Search interface defines a basic method to search
//
// Implementation must strictly follow input and output guide
type Search interface {
	// Search using the input map and output a map of result and error
	Search(input map[string]interface{}, outputType string) (output interface{}, err error)
}

// SimpleSearch basic interface
//type SimpleSearch interface {
//	SimpleSearch(queryInput string, offset int) (
//		total uint64,
//		queryTime time.Duration,
//		result []interface{},
//		err error)
//}

// IdMatch structure contains Id and Score
// of the matched object
//type IdMatch struct {
//	Id    string  `json:"id"`
//	Score float64 `json:"score"`
//}

// FindId basic interface
//type FindId interface {
//	FindId(queryInput string, offset int) (
//		total uint64,
//		queryTime time.Duration,
//		result []*IdMatch,
//		err error)
//}

// ComplexSearch basic interface
//type ComplexSearch interface {
//	ComplexSearch(
//		queryInput string,
//		fields []string,
//		sortBy []string,
//		queryType string,
//		offset int, limit int) (
//		total uint64,
//		queryTime time.Duration,
//		result []interface{},
//		err error)
//
//	BleveComplexSearch(
//		queryInput string,
//		fields []string,
//		sortBy []string,
//		queryType string,
//		offset int, limit int) (*bleve.SearchResult, error)
//}

// BleveSearch basic interface
type BleveSearch interface {
	BleveSearch(req *bleve.SearchRequest) (*bleve.SearchResult, error)

	BleveSearchInContext(
		ctx context.Context,
		req *bleve.SearchRequest) (*bleve.SearchResult, error)
}

//type FacetInput struct {
//	FacetName  string `json:"facetName"`
//	QueryInput string `json:"queryInput"`
//	FacetLimit int    `json:"facetLimit"`
//}
//
//type FacetOutput struct {
//	TermName  string `json:"termName"`
//	TermCount int    `json:"termCount"`
//}

// FacetSearch basic interface
//type FacetSearch interface {
//	FacetSearch(facetInput []map[string]interface{}) (
//		queryTime time.Duration,
//		data map[string][]map[string]interface{},
//		err error)
//}
