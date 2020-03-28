package dodod

type Dodod interface {
	Query
	Mutation

	Search
	FindId

	FacetSearch
	SimpleSearch
	ComplexSearch
	BleveSearch

	DocumentRegistry
}
