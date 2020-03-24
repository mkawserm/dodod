package dodod

import (
	"fmt"
	"testing"
	"time"
)

type mockDocumentStruct struct {
	Id        string `json:"id"`
	DododType string `json:"dododType"`

	CreatedAt uint64 `json:"created_at"`
	UpdatedAt uint64 `json:"updated_at"`

	Duration time.Duration `json:"duration"`
	Complex  complex128    `json:"complex"`
	MyType   struct{}      `json:"my_type"`
}

func (m *mockDocumentStruct) Type() string {
	return "mockDocumentStruct"
}

func (m *mockDocumentStruct) GetId() string {
	return m.Id
}

func TestExtractFields(t *testing.T) {
	t.Helper()
	fmt.Println(ExtractFields(&mockDocumentStruct{}))
	m := &mockDocumentStruct{Id: "1000"}

	fmt.Println(GetId(m))

}
