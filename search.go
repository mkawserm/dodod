package dodod

// Search interface defines a basic method to search
//
// Implementation must strictly follow input and output guide
type Search interface {
	// Search using the input map and output result (based on outputType) and error
	Search(input map[string]interface{}, outputType string) (output interface{}, err error)
}
