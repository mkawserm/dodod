package dodod

type Query interface {
	Read(data []interface{}) (uint64, error)
	ReadUsingId(data []string) (uint64, []interface{}, error)
}
