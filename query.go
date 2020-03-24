package dodod

type Query interface {
	Read(data []Document) (uint64, error)
}
