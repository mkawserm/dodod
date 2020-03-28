package dodod

// Query interface defines all query related methods
type Query interface {
	// Read data using the provided id
	Read(data []string) (uint64, []interface{}, error)

	// GetDocument will fill up the data provided by the interface
	GetDocument(data []interface{}) (uint64, error)

	// GetDocumentWithError will get documents from the database
	// using ids provided in the data string slice
	GetDocumentWithError(data []string) (uint64, []interface{}, error)
}
