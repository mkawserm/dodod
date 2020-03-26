package dodod

type Dodod interface {
	Query
	Mutation

	Search
	FacetSearch
	ComplexSearch

	DocumentRegistry
}
