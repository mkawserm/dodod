package dodod

type Mutation interface {
	Create(data []interface{}) error
	Update(data []interface{}) error
	Delete(data []interface{}) error

	CreateIndex(data []interface{}) error
	UpdateIndex(data []interface{}) error
	DeleteIndex(data []interface{}) error

	CreateDocument(data []interface{}) error
	UpdateDocument(data []interface{}) error
	DeleteDocument(data []interface{}) error
}
