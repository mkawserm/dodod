package dodod

import (
	"reflect"
	"testing"
)

type mockDocumentStruct struct {
	Id        string `json:"id"`
	DododType string `json:"dododType"`

	MyNumber  int64  `json:"my_number"`
	MyNumber1 uint64 `json:"my_number1"`

	MyNumber2 int  `json:"my_number2"`
	MyNumber3 uint `json:"my_number3"`
}

func (m *mockDocumentStruct) Type() string {
	return "mockDocumentStruct"
}

func (m *mockDocumentStruct) GetId() string {
	return m.Id
}

func TestExtractFields(t *testing.T) {
	t.Helper()
	if !reflect.DeepEqual(ExtractFields(&mockDocumentStruct{}), map[string]string{
		"id":         "string",
		"dododType":  "string",
		"my_number":  "int64",
		"my_number1": "uint64",
		"my_number2": "int",
		"my_number3": "uint",
	}) {
		t.Fatalf("reflected fields are not equal")
	}

	m := &mockDocumentStruct{Id: "1000"}

	if GetId(m) != "1000" {
		t.Fatalf("Id does not match")
	}
}

func TestGetType(t *testing.T) {
	t.Helper()

	if GetType(&mockDocumentStruct{}) != "mockDocumentStruct" {
		t.Fatalf("unexpected document")
	}

	if GetType([]string{"1"}) != "" {
		t.Fatalf("unexpected document")
	}
}
