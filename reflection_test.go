package dodod

import (
	"github.com/blevesearch/bleve"
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

type mockDocumentStruct2 struct {
	Id string `json:"id2"`
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

	m2 := &mockDocumentStruct2{Id: "1"}
	if v := GetId(m2); v != "" {
		t.Fatalf("Id should be empty but found %v", v)
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

type mockClassifier struct {
}

func (m *mockClassifier) Type() string {
	return "mockClassifier"
}

type mockCompleteStruct2 struct {
	Id     string `json:"id"`
	Field8 string `json:"field_8"`
}

func (m *mockCompleteStruct2) Type() string {
	return "mockCompleteStruct2"
}

type mockCompleteStruct1 struct {
	Id             string               `json:"id"`
	Field8         string               `json:"field_8"`
	AnotherStruct2 *mockCompleteStruct2 `json:"another_struct_2"`
}

func (m *mockCompleteStruct1) Type() string {
	return "mockCompleteStruct1"
}

type mockCompleteStruct struct {
	Id     string `json:"id"`
	Field0 string `bleve:"-"`
	Field1 string `bleve:"field1"`
	Field2 int    `bleve:"field2"`

	Field4 struct {
		Name string `json:"name"`
	}

	Field5 *mockClassifier `json:"field_5"`
	Field6 mockClassifier  `json:"field_6"`
	Meta   string          `bleve:"meta,index:false,include_term_vectors:false,include_in_all:false"`

	AnotherStruct1 *mockCompleteStruct1 `bleve:"another_struct_1"`
}

func (m *mockCompleteStruct) Type() string {
	return "mockCompleteStruct"
}

type mockErrorStruct1 struct {
	Id    string `json:"id"`
	Meta  string `bleve:"meta:smile1e"`
	Meta1 string `bleve:"meta1,index:false,include_term_vectors:12312,include_in_all:false"`
}

func (m *mockErrorStruct1) Type() string {
	return "mockErrorStruct1"
}

type mockCoverStruct1 struct {
	Id       string `json:"id"`
	T1       string `bleve:"-"`
	T2       string `bleve:"t2,analyzer:english"`
	Meta     string `bleve:"meta,store:false,index:false,include_term_vectors:false,include_in_all:false,doc_values:false"`
	Meta1    string `bleve:"meta1,index1:false,include_in_all:false"`
	Met2     string `bleve:"met2,index:false,include_in_all:false"`
	Location string `bleve:"location,geo_hash:true,index:true,store:true,include_term_vectors:true,include_in_all:true," json:"location"`
}

func (m *mockCoverStruct1) Type() string {
	return "mockCoverStruct1"
}

type mockCoverStruct2 struct {
	Id       string `json:"id"`
	T1       string `bleve:"-"`
	T2       string `bleve:"t2,analyzer:english"`
	Meta     string `bleve:"meta,store:false,index:false,include_term_vectors:false,include_in_all:false,doc_values:false"`
	Meta1    string `bleve:"meta1,index1:false,include_in_all:false"`
	Met2     string `bleve:"met2,index:false,include_in_all:false"`
	Location string `bleve:"location,geo_hash:true,index:true,store:true,include_term_vectors:true,include_in_all:true," json:"location"`

	CustomStruct *mockCoverStruct1 `json:"custom_struct"`

	CustomStruct2 struct {
		Name string `json:"name"`
	} `json:"custom_struct_2"`
}

func (m *mockCoverStruct2) Type() string {
	return "mockCoverStruct2"
}

func Test_registerDocumentMapping(t *testing.T) {
	t.Helper()
	b := make(map[string]interface{})
	if err := registerDocumentMapping(b, &mockClassifier{}); err != ErrUnknownBaseType {
		t.Errorf("unexpected error: %v", err)
	}

	//b1 := &mockClassifier{}
	//if err := registerDocumentMapping(b1, &mockClassifier{}); err !=ErrFieldTypeMismatch {
	//	t.Errorf("unexpected error: %v", err)
	//}

	document := &mockCompleteStruct{}

	if err := registerDocumentMapping(bleve.NewIndexMapping(), document); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := registerDocumentMapping(bleve.NewIndexMapping(), &mockErrorStruct1{}); err != ErrNonBooleanValueForBooleanField {
		t.Errorf("unexpected error: %v", err)
	}

	m := bleve.NewIndexMapping()
	if err := registerDocumentMapping(m, &mockCoverStruct2{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	//d, _ :=json.Marshal(m)
	//t.Errorf(string(d))
}
