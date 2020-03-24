package dodod

import "github.com/blevesearch/bleve/mapping"

type Document interface {
	mapping.Classifier
	GetId() string
}
