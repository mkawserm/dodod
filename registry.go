package dodod

type ModelRegistry interface {
	RegisterModel(model interface{}) error
	GetRegisteredFields() []string
}
