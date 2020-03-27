package dodod

type Dodod interface {
	Query
	Mutation

	Search
	FindId

	FacetSearch
	ComplexSearch

	DocumentRegistry
}
