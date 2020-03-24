package dodod

type DocumentRegistry interface {
	RegisterDocument(model Document) error
	GetRegisteredFields() []string
}
