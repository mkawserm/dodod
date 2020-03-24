package dodod

type Mutation interface {
	Create(data []Document) error
	Update(data []Document) error
	Delete(data []Document) error
}
