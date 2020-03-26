package dodod

type DocumentRegistry interface {
	RegisterDocument(document interface{}) error
	GetRegisteredFields() []string
}
