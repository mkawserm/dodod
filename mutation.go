package dodod

// Mutation interface defines all mutation related methods
type Mutation interface {
	// Create data into the database and index store
	Create(data []interface{}) error
	// Update data inside the database and index store
	Update(data []interface{}) error
	// Delete data from the database and index store
	Delete(data []interface{}) error

	// CreateIndex into the index store
	CreateIndex(data []interface{}) error
	// UpdateIndex inside the index store
	UpdateIndex(data []interface{}) error
	// DeleteIndex from the index store
	DeleteIndex(data []interface{}) error

	// CreateDocument into the database
	CreateDocument(data []interface{}) error
	// UpdateDocument inside the database
	UpdateDocument(data []interface{}) error
	// DeleteDocument from the database
	DeleteDocument(data []interface{}) error
}
