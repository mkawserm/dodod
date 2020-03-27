package dodod

type Query interface {
	Read(data []string) (uint64, []interface{}, error)

	GetDocument(data []interface{}) (uint64, error)
	GetDocumentWithError(data []string) (uint64, []interface{}, error)
}
