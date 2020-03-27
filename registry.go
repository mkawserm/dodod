package dodod

type DocumentRegistry interface {
	RegisterDocument(document interface{}) error
	GetRegisteredFields() []string
	GetRegisteredDocument() map[string]interface{}
}
