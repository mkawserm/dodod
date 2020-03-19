package dodod

type Mutation interface {
	Create(data []interface{}) error
	Update(data []interface{}) error
	Delete(data []interface{}) error
}

type Query interface {
	Read(data []interface{}) error
}

type Search interface {
	Search(filter string, limit int) PageIterator
}

type PageIterator interface {
	HasNextPage() bool
	NextPage()

	TotalPage() uint64
	TotalResults() uint64

	CurrentPageNumber() uint64
	CurrentPageResults() []interface{}
}
