package dodod

type Search interface {
	Search(query string, from uint64) (total uint64, result []Document, err error)
}

type ComplexSearch interface {
	ComplexSearch(query, sortBy, queryType string,
		from uint64, limit uint64) (total uint64, result []Document, err error)
}
