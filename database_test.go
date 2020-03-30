package dodod

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mkawserm/pasap"
	"os"
	"sort"
	"testing"
	"time"
)

import _ "github.com/blevesearch/bleve/analysis/analyzer/keyword"

type DbCredentialsBasic struct {
	Path     string
	Password string
}

func cleanupDb(t *testing.T, path string) {
	t.Helper()
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDb_OpenCloseWithPassword(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := "password"
	credentials := &DbCredentialsBasic{
		Password: dbPassword,
		Path:     dbPath,
	}

	db := &Database{}
	db.SetDbPassword(credentials.Password)
	db.SetDbPath(credentials.Path)
	db.SetupDefaults()
	err := db.Open()

	if err != nil {
		t.Errorf("error occured while opening, error: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("error occured while closing, error: %v", err)
	}

	db = &Database{}
	db.SetDbPassword(credentials.Password)
	db.SetDbPath(credentials.Path)
	db.SetupDefaults()
	err = db.Open()

	if err != nil {
		t.Errorf("error occured while opening, error: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("error occured while closing, error: %v", err)
	}

	cleanupDb(t, dbPath)
}

func TestDb_OpenCloseWithoutPassword(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""
	credentials := &DbCredentialsBasic{
		Password: dbPassword,
		Path:     dbPath,
	}

	{
		db := &Database{}
		db.SetupDefaults()
		db.SetDbPassword(credentials.Password)
		db.SetDbPath(credentials.Path)
		err := db.Open()

		if err != nil {
			t.Errorf("error occured while opening, error: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Errorf("error occured while closing, error: %v", err)
		}
	}

	{
		db := &Database{}
		db.SetDbPassword(credentials.Password)
		db.SetDbPath(credentials.Path)
		db.SetupDefaults()
		err := db.Open()

		if err != nil {
			t.Errorf("error occured while opening, error: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Errorf("error occured while closing, error: %v", err)
		}
	}

	cleanupDb(t, dbPath)
}

func TestDb_Setup(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""
	credentials := &DbCredentialsBasic{
		Password: dbPassword,
		Path:     dbPath,
	}

	t.Run("Call Setup", func(t *testing.T) {
		db := &Database{}
		defer cleanupDb(t, dbPath)
		db.SetDbPassword(credentials.Password)
		db.SetDbPath(credentials.Path)
		db.Setup(
			pasap.NewArgon2idHasher(),
			&pasap.ByteBasedEncoderCredentials{},
			&pasap.ByteBasedVerifierCredentials{},
			&BleveIndexOpener{})
		err := db.Open()

		if err != nil {
			t.Errorf("error occured while opening, error: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Errorf("error occured while closing, error: %v", err)
		}
	})

	t.Run("Call Individual set", func(t *testing.T) {
		db := &Database{}
		defer cleanupDb(t, dbPath)
		db.SetDbPassword(credentials.Password)
		db.SetDbPath(credentials.Path)
		db.SetPasswordHasher(pasap.NewArgon2idHasher())
		db.SetEncoderCredentialsRW(&pasap.ByteBasedEncoderCredentials{})
		db.SetVerifierCredentialsRW(&pasap.ByteBasedVerifierCredentials{})
		db.SetIndexOpener(&BleveIndexOpener{})
		//db.SetIndexMapping(bleve.NewIndexMapping())
		err := db.Open()

		if err != nil {
			t.Errorf("error occured while opening, error: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Errorf("error occured while closing, error: %v", err)
		}
	})
}

/* Mocking failure */
var mockErrOpenIndex = errors.New("mock: failed to open internalIndex")

type mockIndexOpener struct {
}

func (b *mockIndexOpener) BleveIndex(string, *mapping.IndexMappingImpl, string, map[string]interface{}) (bleve.Index, error) {
	return nil, mockErrOpenIndex
}

//var mockErrInvalidPath = errors.New("mock: invalid path")
//var mockErrInvalidPass = errors.New("moc: invalid password")
//
//type mockCredentialsInvalidPath struct {
//}
//
//func (d *mockCredentialsInvalidPath) ReadPath() (dbPath string, err error) {
//	return "", mockErrInvalidPath
//}
//
//func (d *mockCredentialsInvalidPath) ReadPassword() (password string, err error) {
//	return "123123", nil
//}

//type mockCredentialsInvalidPassword struct {
//}
//
//func (d *mockCredentialsInvalidPassword) ReadPath() (dbPath string, err error) {
//	return "/tmp/test_db", nil
//}
//
//func (d *mockCredentialsInvalidPassword) ReadPassword() (password string, err error) {
//	return "", mockErrInvalidPass
//}

func TestMockOpenFailure(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""
	credentials := &DbCredentialsBasic{
		Password: dbPassword,
		Path:     dbPath,
	}

	t.Run("Open internalIndex failure", func(t *testing.T) {
		db := &Database{}
		defer cleanupDb(t, dbPath)
		db.SetupDefaults()
		db.SetDbPassword(credentials.Password)
		db.SetDbPath(credentials.Path)
		db.SetIndexOpener(&mockIndexOpener{})
		err := db.Open()
		if err != mockErrOpenIndex {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("error occured while closing, error: %v", err)
		}
	})

	t.Run("Credentials path failure", func(t *testing.T) {
		db := &Database{}
		db.SetupDefaults()
		err := db.Open()
		if err != ErrEmptyPath {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("error occured while closing, error: %v", err)
		}
	})

	//t.Run("Credentials password failure", func(t *testing.T) {
	//	db := &Database{}
	//	db.SetupDefaults()
	//	db.SetDbCredentials(&mockCredentialsInvalidPassword{})
	//	err := db.Open()
	//	if err != mockErrInvalidPass {
	//		t.Fatalf("unexpected error: %v", err)
	//	}
	//	if err := db.Close(); err != nil {
	//		t.Fatalf("error occurred while closing, error: %v", err)
	//	}
	//})

}

func TestDatabase_ChangePassword(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := "password"
	dbNewPassword := "password2"

	defer cleanupDb(t, dbPath)

	db := &Database{}
	//db.SetIndexStoreName("scorch")
	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}

	db2 := &Database{}
	//db2.SetIndexStoreName("scorch")
	db2.SetupDefaults()
	db2.SetDbPassword(dbPassword)
	db2.SetDbPath(dbPath)

	err = db2.ChangePassword(dbNewPassword)
	if err != nil {
		t.Fatalf("unexpected error while changing password: %v", err)
	}

	db2.SetDbPassword(dbNewPassword)

	err = db2.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db2.Close(); err != nil {
		t.Fatalf("error occurred while closing, error: %v", err)
	}
}

func TestDatabase_ChangePassword2(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := "password"
	dbNewPassword := "password2"

	defer cleanupDb(t, dbPath)

	db := &Database{}
	db.SetIndexStoreName("scorch")
	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}

	db2 := &Database{}
	db2.SetIndexStoreName("scorch")
	db2.SetupDefaults()
	db2.SetDbPassword(dbPassword)
	db2.SetDbPath(dbPath)

	err = db2.ChangePassword(dbNewPassword)
	if err != nil {
		t.Fatalf("unexpected error while changing password: %v", err)
	}

	db2.SetDbPassword(dbNewPassword)

	err = db2.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db2.Close(); err != nil {
		t.Fatalf("error occurred while closing, error: %v", err)
	}
}

type MyTestDocument struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (m *MyTestDocument) Type() string {
	return "MyTestDocument"
}

func (m *MyTestDocument) GetId() string {
	return m.Id
}

type MyTestDocument2 struct {
	Id   string `json:"id"`
	Name int    `json:"name"`
}

func (m *MyTestDocument2) Type() string {
	return "MyTestDocument2"
}

func (m *MyTestDocument2) GetId() string {
	return m.Id
}

func TestDatabase_RegisterDocument(t *testing.T) {
	t.Helper()

	db := &Database{}
	if err := db.RegisterDocument(&MyTestDocument{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.RegisterDocument(&MyTestDocument{}); err != ErrDocumentTypeAlreadyRegistered {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.RegisterDocument(&MyTestDocument2{}); err != ErrFieldTypeMismatch {
		t.Fatalf("unexpected error: %v", err)
	}

	data := db.GetRegisteredFields()
	sort.Strings(data)
	n := len(data)

	if n != 2 {
		t.Fatalf("Registered fields are not equal")
	}

	if sort.SearchStrings(data, "id") > n {
		t.Fatalf("id field not found")
	}
	if sort.SearchStrings(data, "name") > n {
		t.Fatalf("name field not found")
	}
}

func TestDatabase_IsDatabaseReady(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""
	defer cleanupDb(t, dbPath)

	db := &Database{}
	if db.IsDatabaseReady() {
		t.Fatalf("database should not be ready")
	}

	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !db.IsDatabaseReady() {
		t.Fatalf("database should be ready")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

func TestDatabase_Create(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""
	defer cleanupDb(t, dbPath)

	db := &Database{}
	if err := db.RegisterDocument(&MyTestDocument{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db.IsDatabaseReady() {
		t.Fatalf("database should not be ready")
	}

	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	if err := db.Create([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !db.IsDatabaseReady() {
		t.Fatalf("database should be ready")
	}

	if err := db.Create([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	//if err := db.Create([]Document{&MyTestDocument{
	//	Id:   "1",
	//	Name: "Test1",
	//}}); err !=ErrDatabaseTransactionFailed {
	//	t.Fatalf("unexpected error: %v", err)
	//}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

func TestDatabase_CRUD(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""

	defer cleanupDb(t, dbPath)

	db := &Database{}
	if err := db.RegisterDocument(&MyTestDocument{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db.IsDatabaseReady() {
		t.Fatalf("database should not be ready")
	}

	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	if err := db.Create([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := db.GetDocument([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, _, err := db.Read([]string{"1"}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := db.GetDocumentWithError([]string{"1"}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.Update([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.Delete([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	// open
	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !db.IsDatabaseReady() {
		t.Fatalf("database should be ready")
	}

	if err := db.Create([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []interface{}{&MyTestDocument{Id: "1"}}

	if n, err := db.GetDocument(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
	}

	if n, data, err := db.Read([]string{"1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
		if len(data) != 1 {
			t.Fatalf("read failure")
		}
	}

	if n, data, err := db.Read([]string{"1", "2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
		if len(data) != 1 {
			t.Fatalf("read failure")
		}
	}

	if n, data, err := db.GetDocumentWithError([]string{"1", "2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
		if len(data) != 2 {
			t.Fatalf("read failure")
		}
	}

	if err := db.Update([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "UpdatedTest1",
	}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n, err := db.GetDocument(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}

		if val, ok := data[0].(*MyTestDocument); !ok {
			t.Fatalf("document conversion failed")
		} else {
			if val.Name != "UpdatedTest1" {
				t.Fatalf("document update failed")
			}
		}
	}

	if err := db.Delete(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n, err := db.GetDocument(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 0 {
			t.Fatalf("delete failure")
		}
	}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

func TestDatabase_DOCUMENT_CRUD(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""

	defer cleanupDb(t, dbPath)

	db := &Database{}
	if err := db.RegisterDocument(&MyTestDocument{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db.IsDatabaseReady() {
		t.Fatalf("database should not be ready")
	}

	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	if err := db.CreateDocument([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.UpdateDocument([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.DeleteDocument([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	// open
	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !db.IsDatabaseReady() {
		t.Fatalf("database should be ready")
	}

	if err := db.CreateDocument([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []interface{}{&MyTestDocument{Id: "1"}}

	if n, err := db.GetDocument(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
	}

	if err := db.UpdateDocument([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "UpdatedTest1",
	}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n, err := db.GetDocument(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}

		if val, ok := data[0].(*MyTestDocument); !ok {
			t.Fatalf("document conversion failed")
		} else {
			if val.Name != "UpdatedTest1" {
				t.Fatalf("document update failed")
			}
		}
	}

	if err := db.DeleteDocument(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n, err := db.GetDocument(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 0 {
			t.Fatalf("delete failure")
		}
	}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

func TestDatabase_CreateIndex(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""

	defer cleanupDb(t, dbPath)

	db := &Database{}
	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	data := []interface{}{&MyTestDocument{Id: "1"}}

	if err := db.CreateIndex(data); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.CreateIndex([]interface{}{&map[string]string{"id": "1"}}); err != ErrInvalidDocument {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := db.CreateIndex([]interface{}{&MyTestDocument{Id: ""}}); err != ErrIdCanNotBeEmpty {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.CreateIndex(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatabase_UpdateIndex(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""

	defer cleanupDb(t, dbPath)

	db := &Database{}
	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	data := []interface{}{&MyTestDocument{Id: "1"}}

	if err := db.UpdateIndex(data); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.UpdateIndex([]interface{}{&map[string]string{"id": "1"}}); err != ErrInvalidDocument {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := db.UpdateIndex([]interface{}{&MyTestDocument{Id: ""}}); err != ErrIdCanNotBeEmpty {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.UpdateIndex(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatabase_DeleteIndex(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""

	defer cleanupDb(t, dbPath)

	db := &Database{}
	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	data := []interface{}{&MyTestDocument{Id: "1"}}

	if err := db.DeleteIndex(data); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.DeleteIndex([]interface{}{&map[string]string{"id": "1"}}); err != ErrInvalidDocument {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := db.DeleteIndex([]interface{}{&MyTestDocument{Id: ""}}); err != ErrIdCanNotBeEmpty {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := db.DeleteIndex(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type CustomDocument struct {
	Id           string `json:"id"`
	CustomField1 string `json:"custom_field_1"`
	CustomField2 string `json:"custom_field_2"`
}

func (c *CustomDocument) GetId() string {
	return c.Id
}

func (c *CustomDocument) Type() string {
	return "CustomDocument"
}

func TestDatabase_Search(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""

	defer cleanupDb(t, dbPath)

	db := &Database{}

	if err := db.RegisterDocument(&CustomDocument{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	addData := []interface{}{&CustomDocument{
		Id:           "1",
		CustomField1: "1 value field 1",
		CustomField2: "1 value field 2",
	},
		&CustomDocument{
			Id:           "2",
			CustomField1: "2 value field 1",
			CustomField2: "2 value field 2",
		},
		&CustomDocument{
			Id:           "3",
			CustomField1: "3 value field 1",
			CustomField2: "3 value field 2",
		},
		&CustomDocument{
			Id:           "4",
			CustomField1: "4 value field 1",
			CustomField2: "4 value field 2",
		},
		&CustomDocument{
			Id:           "5",
			CustomField1: "5 value field 1",
			CustomField2: "5 value field 2",
		},
		&CustomDocument{
			Id:           "6",
			CustomField1: "6 value field 1",
			CustomField2: "6 value field 2",
		},
		&CustomDocument{
			Id:           "7",
			CustomField1: "7 value field 1",
			CustomField2: "7 value field 2",
		},
		&CustomDocument{
			Id:           "8",
			CustomField1: "8 value field 1",
			CustomField2: "8 value field 2",
		},
		&CustomDocument{
			Id:           "9",
			CustomField1: "9 value field 1",
			CustomField2: "9 value field 2",
		},
		&CustomDocument{
			Id:           "10",
			CustomField1: "10 value field 1",
			CustomField2: "10 value field 2",
		},
		&CustomDocument{
			Id:           "11",
			CustomField1: "11 value field 1",
			CustomField2: "11 value field 2",
		},
		&CustomDocument{
			Id:           "12",
			CustomField1: "12 value field 1",
			CustomField2: "12 value field 2",
		},
		&CustomDocument{
			Id:           "13",
			CustomField1: "13 value field 1",
			CustomField2: "13 value field 2",
		},
		&CustomDocument{
			Id:           "14",
			CustomField1: "14 value field 1",
			CustomField2: "14 value field 2",
		},
	}

	if err := db.Create(addData); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db.SetSearchResultLimit(5)

	// Now search data

	//if total, _, result, err := db.SimpleSearch("value", 0); err != nil {
	//	t.Fatalf("unexpected error: %v", err)
	//} else {
	//	if len(result) != 5 {
	//		t.Fatalf("result should be 5 but got %v", len(result))
	//	}
	//	if total != 14 {
	//		t.Fatalf("total result should be 14 but got %v", total)
	//	}
	//	doc1 := result[0].(*CustomDocument)
	//	doc2 := result[1].(*CustomDocument)
	//	if doc1.Id == doc2.Id {
	//		t.Fatalf("Document id should not be equal")
	//	}
	//	if doc1.CustomField1 == doc2.CustomField1 {
	//		t.Fatalf("field should not be equal")
	//	}
	//	if doc1.CustomField2 == doc2.CustomField2 {
	//		t.Fatalf("field should not be equal")
	//	}
	//	//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	//}

	//if total, _, result, err := db.SimpleSearch("value", 5); err != nil {
	//	t.Fatalf("unexpected error: %v", err)
	//} else {
	//	if len(result) != 5 {
	//		t.Fatalf("result should be 5 but got %v", len(result))
	//	}
	//	if total != 14 {
	//		t.Fatalf("total result should be 14 but got %v", total)
	//	}
	//	doc1 := result[0].(*CustomDocument)
	//	doc2 := result[1].(*CustomDocument)
	//	if doc1.Id == doc2.Id {
	//		t.Fatalf("Document id should not be equal")
	//	}
	//	if doc1.CustomField1 == doc2.CustomField1 {
	//		t.Fatalf("field should not be equal")
	//	}
	//	if doc1.CustomField2 == doc2.CustomField2 {
	//		t.Fatalf("field should not be equal")
	//	}
	//	//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	//}

	//if total, _, result, err := db.SimpleSearch("value", 10); err != nil {
	//	t.Fatalf("unexpected error: %v", err)
	//} else {
	//	if len(result) != 4 {
	//		t.Fatalf("result should be 4 but got %v", len(result))
	//	}
	//	if total != 14 {
	//		t.Fatalf("total result should be 14 but got %v", total)
	//	}
	//	doc1 := result[0].(*CustomDocument)
	//	doc2 := result[1].(*CustomDocument)
	//	if doc1.Id == doc2.Id {
	//		t.Fatalf("Document id should not be equal")
	//	}
	//	if doc1.CustomField1 == doc2.CustomField1 {
	//		t.Fatalf("field should not be equal")
	//	}
	//	if doc1.CustomField2 == doc2.CustomField2 {
	//		t.Fatalf("field should not be equal")
	//	}
	//	//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	//}

	//if total, _, result, err := db.SimpleSearch("value", 15); err != nil {
	//	t.Fatalf("unexpected error: %v", err)
	//} else {
	//	if len(result) != 0 {
	//		t.Fatalf("result should be 0 but got %v", len(result))
	//	}
	//	if total != 14 {
	//		t.Fatalf("total result should be 14 but got %v", total)
	//	}
	//	//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	//}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

//func TestDatabase_ComplexSearch(t *testing.T) {
//	t.Helper()
//
//	dbPath := "/tmp/dodod"
//	dbPassword := ""
//
//	defer cleanupDb(t, dbPath)
//
//	db := &Database{}
//
//	if err := db.RegisterDocument(&CustomDocument{}); err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	db.SetupDefaults()
//	db.SetDbPassword(dbPassword)
//	db.SetDbPath(dbPath)
//
//	err := db.Open()
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	addData := []interface{}{&CustomDocument{
//		Id:           "1",
//		CustomField1: "1 value field 1",
//		CustomField2: "1 value field 2",
//	},
//		&CustomDocument{
//			Id:           "2",
//			CustomField1: "2 value field 1",
//			CustomField2: "2 value field 2",
//		},
//		&CustomDocument{
//			Id:           "3",
//			CustomField1: "3 value field 1",
//			CustomField2: "3 value field 2",
//		},
//		&CustomDocument{
//			Id:           "4",
//			CustomField1: "4 value field 1",
//			CustomField2: "4 value field 2",
//		},
//		&CustomDocument{
//			Id:           "5",
//			CustomField1: "5 value field 1",
//			CustomField2: "5 value field 2",
//		},
//		&CustomDocument{
//			Id:           "6",
//			CustomField1: "6 value field 1",
//			CustomField2: "6 value field 2",
//		},
//		&CustomDocument{
//			Id:           "7",
//			CustomField1: "7 value field 1",
//			CustomField2: "7 value field 2",
//		},
//		&CustomDocument{
//			Id:           "8",
//			CustomField1: "8 value field 1",
//			CustomField2: "8 value field 2",
//		},
//		&CustomDocument{
//			Id:           "9",
//			CustomField1: "9 value field 1",
//			CustomField2: "9 value field 2",
//		},
//		&CustomDocument{
//			Id:           "10",
//			CustomField1: "10 value field 1",
//			CustomField2: "10 value field 2",
//		},
//		&CustomDocument{
//			Id:           "11",
//			CustomField1: "11 value field 1",
//			CustomField2: "11 value field 2",
//		},
//		&CustomDocument{
//			Id:           "12",
//			CustomField1: "12 value field 1",
//			CustomField2: "12 value field 2",
//		},
//		&CustomDocument{
//			Id:           "13",
//			CustomField1: "13 value field 1",
//			CustomField2: "13 value field 2",
//		},
//		&CustomDocument{
//			Id:           "14",
//			CustomField1: "14 value field 1",
//			CustomField2: "14 value field 2",
//		},
//	}
//
//	if err := db.Create(addData); err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//
//	db.SetSearchResultLimit(5)
//
//	// Now search data
//
//	//sortBy := []string{"-id"}
//	//queryType := "QueryString"
//	//limit := 10
//	//fields := []string{"*"}
//
//	//if total, _, result, err := db.ComplexSearch("value", fields, sortBy, queryType, 0, limit); err != nil {
//	//	t.Fatalf("unexpected error: %v", err)
//	//} else {
//	//	if len(result) != 10 {
//	//		t.Fatalf("result should be 10 but got %v", len(result))
//	//	}
//	//	if total != 14 {
//	//		t.Fatalf("total result should be 14 but got %v", total)
//	//	}
//	//
//	//	doc1 := result[0].(*CustomDocument)
//	//	doc2 := result[1].(*CustomDocument)
//	//	if doc1.Id == doc2.Id {
//	//		t.Fatalf("Document id should not be equal")
//	//	}
//	//	if doc1.CustomField1 == doc2.CustomField1 {
//	//		t.Fatalf("field should not be equal")
//	//	}
//	//	if doc1.CustomField2 == doc2.CustomField2 {
//	//		t.Fatalf("field should not be equal")
//	//	}
//	//	//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
//	//}
//	//
//	//if total, _, result, err := db.ComplexSearch("value", fields, sortBy, queryType, 5, limit); err != nil {
//	//	t.Fatalf("unexpected error: %v", err)
//	//} else {
//	//	if len(result) != 9 {
//	//		t.Fatalf("result should be 9 but got %v", len(result))
//	//	}
//	//	if total != 14 {
//	//		t.Fatalf("total result should be 14 but got %v", total)
//	//	}
//	//	//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
//	//}
//
//	if err := db.Close(); err != nil {
//		t.Fatalf("error occured while closing, error: %v", err)
//	}
//}

func TestDatabase_EncodeDecodeDocument(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := ""

	defer cleanupDb(t, dbPath)

	db := &Database{}

	if err := db.RegisterDocument(&CustomDocument{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db.SetupDefaults()
	db.SetDbPassword(dbPassword)
	db.SetDbPath(dbPath)

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := &CustomDocument{
		Id:           "1",
		CustomField1: "field 1",
		CustomField2: "field 2",
	}

	if d, err := db.EncodeDocument(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if _, err := db.DecodeDocument(d); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			//fmt.Println(n.(*CustomDocument))
		}
	}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

type FirstMixedType struct {
	Id        string `json:"id"`
	MixedType string `json:"mixed_type"`
	Location  string `bleve:"location,geo_hash:true,index:true,store:true,include_in_all:true" json:"location"`

	Field1 string    `json:"field_1"`
	Field2 string    `json:"field_2"`
	Field3 int64     `json:"field_3"`
	Field4 float64   `json:"field_4"`
	Field5 bool      `json:"field_5"`
	Field6 time.Time `json:"field_6"`

	Field7  []byte    `json:"field_7"`
	Field8  []string  `json:"field_8"`
	Field9  []int64   `json:"field_9"`
	Field10 []float64 `json:"field_10"`
}

//func (m *FirstMixedType) BleveType() string {
//	return "FirstMixedType"
//}

func (m *FirstMixedType) Type() string {
	return "FirstMixedType"
}

func (m *FirstMixedType) GetId() string {
	return m.Id
}

type SecondMixedType struct {
	Id        string `json:"id"`
	MixedType string `json:"mixed_type"`
	//Location string `json:"location"`

	Field11 string    `json:"field_11"`
	Field12 string    `json:"field_12"`
	Field13 int64     `json:"field_13"`
	Field14 float64   `json:"field_14"`
	Field15 bool      `json:"field_15"`
	Field16 time.Time `json:"field_16"`

	Field17 []byte    `json:"field_17"`
	Field18 []string  `json:"field_18"`
	Field19 []int64   `json:"field_19"`
	Field20 []float64 `json:"field_20"`
}

func (m *SecondMixedType) Type() string {
	return "SecondMixedType"
}

func (m *SecondMixedType) GetId() string {
	return m.Id
}

type ThirdMixedType struct {
	Id        string `json:"id"`
	MixedType string `json:"mixed_type"`
	//Location string `json:"location"`

	Field21 string    `json:"field_21"`
	Field22 string    `json:"field_22"`
	Field23 int64     `json:"field_23"`
	Field24 float64   `json:"field_24"`
	Field25 bool      `json:"field_25"`
	Field26 time.Time `json:"field_26"`

	Field27 []byte    `json:"field_27"`
	Field28 []string  `json:"field_28"`
	Field29 []int64   `json:"field_29"`
	Field30 []float64 `json:"field_30"`
}

func (m *ThirdMixedType) Type() string {
	return "ThirdMixedType"
}

func (m *ThirdMixedType) GetId() string {
	return m.Id
}

func createTestData() (documentTypes []interface{}, testIds []string, testData []interface{}) {
	documentTypes = []interface{}{
		&FirstMixedType{},
		&SecondMixedType{},
		&ThirdMixedType{},
	}

	testIds = []string{"1", "2", "3"}

	testData = []interface{}{
		&FirstMixedType{
			Id:        "1",
			MixedType: "FirstMixedType",
			Location:  "wecpjc2b27ev",
			Field1:    "FMTF 1",
			Field2:    "FMTF 2",
			Field3:    13,
			Field4:    14.1,
			Field5:    true,
			Field6:    time.Now(),
			Field7:    []byte{1, 2, 3, 4, 5},
			Field8:    []string{"1", "2", "3"},
			Field9:    []int64{111111111111111, 211111111111122, 311111111111133},
			Field10:   []float64{1111111111111.11, 2111111111111.22, 3111111111111.33},
		},
		&SecondMixedType{
			Id:        "2",
			MixedType: "SecondMixedType",
			//Location: "wecpkbeddsmf",
			Field11: "SMTF 11",
			Field12: "SMTF 12",
			Field13: 213,
			Field14: 214.2,
			Field15: false,
			Field16: time.Now(),
			Field17: []byte{21, 22, 23, 24, 25},
			Field18: []string{"21", "22", "23"},
			Field19: []int64{211111111111111, 311111111111122, 411111111111133},
			Field20: []float64{2111111111111.11, 3111111111111.22, 4111111111111.33},
		},
		&ThirdMixedType{
			Id:        "3",
			MixedType: "ThirdMixedType",
			//Location: "wecnzm94b80h",
			Field21: "TMTF 21",
			Field22: "TMTF 22",
			Field23: 323,
			Field24: 324.3,
			Field25: true,
			Field26: time.Now(),
			Field27: []byte{31, 32, 33, 34, 35},
			Field28: []string{"31", "32", "33"},
			Field29: []int64{311111111111111, 411111111111122, 511111111111133},
			Field30: []float64{3111111111111.11, 4111111111111.22, 5111111111111.33},
		},
	}

	return
}

func TestDatabaseTable(t *testing.T) {
	t.Helper()

	documentTypes, testIds, testData := createTestData()

	dbPath := "/tmp/dodod"
	defer cleanupDb(t, dbPath)

	db := &Database{}
	//db.SetIndexStoreName("scorch")
	db.SetDbPath(dbPath)
	db.SetupDefaults()

	// Test Data
	//for _, v:= range testData {
	//	data, _ := json.Marshal(v)
	//	t.Errorf("%s", data)
	//}

	// Register documents before opening database
	for _, v := range documentTypes {
		if err := db.RegisterDocument(v); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Register document expected failure
	if err := db.RegisterDocument(map[string]string{"1": "1"}); err != ErrInvalidDocument {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get registered document
	if v := db.GetRegisteredDocument(); v == nil {
		t.Fatalf("registered document should not be nil")
	}

	// Encode document expected error
	if _, err := db.EncodeDocument(map[string]string{"1": "1"}); err != ErrInvalidDocument {
		t.Fatalf("unexpected error: %v", err)
	}

	// Decode document expected error
	if _, err := db.DecodeDocument([]byte("1231")); err != ErrInvalidData {
		t.Fatalf("unexpected error: %v", err)
	}

	// Decode document expected error
	if _, err := db.DecodeDocument([]byte("1234567890")); err != ErrInvalidData {
		t.Fatalf("unexpected error: %v", err)
	}

	// Open database
	if err := db.Open(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get internal db for just covering
	if v := db.GetInternalDatabase(); v == nil {
		t.Fatalf("Internal database should not be nil")
	}

	if v := db.GetInternalIndex(); v == nil {
		t.Fatalf("Internal index should not be nil")
	}

	// Add test data to the database
	if err := db.Create(testData); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read data using id
	if total, data, err := db.Read(testIds); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(testIds) != len(data) {
			t.Fatalf("read failure")
		}
		if total != uint64(len(data)) {
			t.Fatalf("read failure")
		}

		if v, ok := data[0].(Document); ok {
			if v.Type() != "FirstMixedType" {
				t.Fatalf("unknown type")
			}
		}

		if v, ok := data[1].(Document); ok {
			if v.Type() != "SecondMixedType" {
				t.Fatalf("unknown type")
			}
		}

		if v, ok := data[2].(Document); ok {
			if v.Type() != "ThirdMixedType" {
				t.Fatalf("unknown type")
			}
		}

		if v, ok := data[0].(*FirstMixedType); ok {
			if v.Field10[0] != 1111111111111.11 {
				t.Fatalf("data should be euqal")
			}
		} else {
			t.Fatalf("FirstMixedType conversion failed")
		}

	}

	// Search
	//input := make(map[string]interface{})
	//input["sort"] = []string{"_id"}
	//input["fields"] = []string{"*"}

	//input["fields"] = []string{"*"}
	//if data, err := db.Search(input, "bytes"); err != nil {
	//	t.Fatalf("unexpected error: %v", err)
	//} else {
	//	if data == nil {
	//		t.Fatalf("data should not be nil")
	//	}
	//	//t.Errorf("%s", data)
	//}

	t.Run("outputType mapIncludeData", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		if data, err := db.Search(input, "mapIncludeData"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			//output, _ := json.Marshal(data)
			//t.Errorf("%v", data)
		}
	})

	t.Run("outputType map", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		if data, err := db.Search(input, "map"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
		}
	})

	t.Run("outputType bytes", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		if data, err := db.Search(input, "bytes"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
		}
	})

	t.Run("outputType bleveSearchResult", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
		}
	})

	t.Run("outputType default/empty", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		if data, err := db.Search(input, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
		}
	})

	t.Run("Search With QueryString With facets", func(t *testing.T) {
		input := make(map[string]interface{})
		input["size"] = 0
		input["facets"] = []interface{}{
			map[string]interface{}{
				"name":  "types",
				"field": "mixed_type",
				"size":  10,
			},
			map[string]interface{}{
				"name":  "Field3",
				"field": "field_3",
				"size":  10,
				"numeric_range": []interface{}{
					map[string]interface{}{
						"name": "Test",
						"min":  1.0,
						"max":  1000.0,
					},
				},
			},
			map[string]interface{}{
				"name":  "Field6",
				"field": "field_6",
				"size":  10,
				"date_time_range": []interface{}{
					map[string]string{
						"name":  "Field6Min",
						"start": "12:00", //format is not correct
						"end":   "6:00",  // format is not correct
					},
				},
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if len(bleveSearchResult.Facets) == 0 {
				t.Fatalf("Facets should not be zero")
			}
		}
	})

	t.Run("Search With QueryString", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["size"] = 100
		input["from"] = 0
		input["fields"] = []string{"*"}
		input["explain"] = true
		input["include_locations"] = true
		input["score"] = "1"

		input["query"] = map[string]interface{}{
			"name": "QueryString",
			"p": map[string]interface{}{
				"q": "id:1",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 1 {
				t.Fatalf("Total Expected 1, but found: %v", bleveSearchResult.Total)
			}

			fmt.Println(bleveSearchResult.Hits[0].Fields)

			//if lat, found := bleveSearchResult.Hits[0].Fields["location"].(string); !found {
			//	t.Fatalf("Location data not found")
			//} else {
			//	if lat != "wecpjc2b27ev" {
			//		t.Fatalf("Lat does not match.")
			//	}
			//}
		}
	})

	t.Run("Search With QueryString Search After", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["size"] = 100
		input["search_after"] = []string{"2"}
		input["query"] = map[string]interface{}{
			"name": "QueryString",
			"p": map[string]interface{}{
				"q": "id:3",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 1 {
				t.Fatalf("Total Expected 1, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With QueryString Search Before", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["size"] = 100
		input["search_before"] = []string{"3"}
		input["query"] = map[string]interface{}{
			"name": "QueryString",
			"p": map[string]interface{}{
				"q": "id:2",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 1 {
				t.Fatalf("Total Expected 1, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With QueryString with highlight option", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["size"] = 100
		input["highlight"] = map[string]interface{}{
			"style":  "html",
			"fields": []string{"id"},
		}

		input["query"] = map[string]interface{}{
			"name": "QueryString",
			"p": map[string]interface{}{
				"q": "id:2",
			},
		}

		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 1 {
				t.Fatalf("Total Expected 1, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With Fuzzy Expected Error", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "Fuzzy",
			"p": map[string]interface{}{
				"term":      "2",
				"fuzziness": 3,
			},
		}
		if _, err := db.Search(input, "bleveSearchResult"); err == nil {
			t.Fatalf("Search should reurn an error")
		}
	})

	t.Run("Search With Match", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "Match",
			"p": map[string]interface{}{
				"match": "SecondMixedType",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 1 {
				t.Fatalf("Total Expected 1, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With MatchPhrase", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "MatchPhrase",
			"p": map[string]interface{}{
				"match_phrase": "SecondMixedType",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 1 {
				t.Fatalf("Total Expected 1, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With Prefix", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "Prefix",
			"p": map[string]interface{}{
				"prefix": "21",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 2 {
				t.Fatalf("Total Expected 2, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With Wildcard", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "Wildcard",
			"p": map[string]interface{}{
				"wildcard": "2?",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 2 {
				t.Fatalf("Total Expected 2, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With Fuzzy", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "Fuzzy",
			"p": map[string]interface{}{
				"term": "2",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 3 {
				t.Fatalf("Total Expected 3, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With Term", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "Term",
			"p": map[string]interface{}{
				"term": "1",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 1 {
				t.Fatalf("Total Expected 1, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	t.Run("Search With Regexp", func(t *testing.T) {
		input := make(map[string]interface{})
		input["sort"] = []string{"_id"}
		input["query"] = map[string]interface{}{
			"name": "Regexp",
			"p": map[string]interface{}{
				"regexp": "2",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)
			if bleveSearchResult.Total != 2 {
				t.Fatalf("Total Expected 2, but found: %v", bleveSearchResult.Total)
			}
		}
	})

	// Close the database
	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

//type mockBleveIndexOpener struct {
//}
//
//func (b *mockBleveIndexOpener) BleveIndex(dbPath string,
//	indexMapping *mapping.IndexMappingImpl,
//	indexName string,
//	config map[string]interface{}) (bleve.Index, error) {
//	return bleve.NewMemOnly(indexMapping)
//}

func TestDatabase_GeoSearch(t *testing.T) {
	t.Helper()

	documentTypes, _, _ := createTestData()

	dbPath := "/tmp/dodod"
	defer cleanupDb(t, dbPath)

	db := &Database{}
	//db.SetIndexStoreName(scorch.Name)
	db.SetDbPath(dbPath)
	db.SetupDefaults()
	//db.SetIndexOpener(&mockBleveIndexOpener{})

	// Test Data
	//for _, v:= range testData {
	//	data, _ := json.Marshal(v)
	//	t.Errorf("%s", data)
	//}

	// Register documents before opening database
	for _, v := range documentTypes {
		if err := db.RegisterDocument(v); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		break
	}

	// Open database
	if err := db.Open(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db.GetIndexMapping()

	testData := []interface{}{
		&FirstMixedType{
			Id:        "1",
			MixedType: "FirstMixedType",
			Location:  "wecpjc2b27ev",
		},
		&FirstMixedType{
			Id:       "2",
			Location: "wecpkbeddsmf",
		},
		&FirstMixedType{
			Id:       "3",
			Location: "wecnzm94b80h",
		},
		&FirstMixedType{
			Id:       "4",
			Location: "wecpk8tne453",
		},
		&FirstMixedType{
			Id:       "5",
			Location: "wecnycjgz1u3",
		},
		&FirstMixedType{
			Id:       "6",
			Location: "wecny57t09cu",
		},
		&FirstMixedType{
			Id:       "7",
			Location: "wecpkb80s09t",
		},
		&FirstMixedType{
			Id:       "8",
			Location: "wecpjbbru3dj",
		},
		&FirstMixedType{
			Id:       "9",
			Location: "wecnznn0hzr1",
		},
		&FirstMixedType{
			Id:       "10",
			Location: "wecpqgeu2uzw",
		},
	}

	// Add test data to the database
	if err := db.Create(testData); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Search With GeoDistance", func(t *testing.T) {

		//lat, lon := 22.371154, 114.112603
		//q := bleve.NewGeoDistanceQuery(lon, lat, "1km")
		//
		//req := bleve.NewSearchRequest(q)
		//req.Fields = []string{"*"}
		//req.SortBy([]string{"_id"})
		//
		//sr, err := db.internalIndex.Search(req)
		//if err != nil {
		//	t.Errorf("unexpected error:%v", err)
		//}
		//if sr.Total != 3 {
		//	t.Errorf("Size expected: 3, actual %d\n", sr.Total)
		//}
		//
		//t.Logf("Total found: %d", sr.Total)

		input := make(map[string]interface{})
		input["fields"] = []string{"*"}
		input["sort"] = []string{"_id"}
		input["size"] = 100
		input["from"] = 0
		input["explain"] = true
		input["include_locations"] = true
		input["query"] = map[string]interface{}{
			"name": "GeoDistance",
			"p": map[string]interface{}{
				"lon":      114.112603,
				"lat":      22.371154,
				"distance": "1km",
				"field":    "location",
			},
		}
		if data, err := db.Search(input, "bleveSearchResult"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			if data == nil {
				t.Fatalf("data should not be nil")
			}
			bleveSearchResult := data.(*bleve.SearchResult)

			//t.Logf("\n\n%s\n\n", bleveSearchResult.Request.Query)
			//indexBytes, _ := json.Marshal(db.internalIndex.Mapping())
			//t.Logf("%s", string(indexBytes))

			if bleveSearchResult.Total <= 0 {
				t.Fatalf("Total Expected at least 1, but found: %v", bleveSearchResult.Total)
			}
			t.Logf("Total found: %d", bleveSearchResult.Total)
		}

		indexBytes, _ := json.Marshal(db.internalIndex.Mapping())
		t.Logf("%s", string(indexBytes))

	})

	// Close the database
	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

/*func TestDatabase_GeoSearchIndexOnly(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	defer cleanupDb(t, dbPath)

	db := &Database{}
	//db.SetIndexStoreName("scorch")
	db.SetDbPath(dbPath)
	db.SetupDefaults()
	db.SetIndexOpener(&mockBleveIndexOpener{})

	db.indexMapping = bleve.NewIndexMapping()
	documentMapping := bleve.NewDocumentMapping()
	documentMapping.AddFieldMappingsAt("location", bleve.NewGeoPointFieldMapping())
	db.indexMapping.AddDocumentMapping("FirstMixedType", documentMapping)

	// Open database
	if err := db.Open(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db.internalIndex.Index("1", &FirstMixedType{
		Id:       "1",
		MixedType: "FirstMixedType",
		Location:  "wecpjc2b27ev",
	})
	db.internalIndex.Index("2", &FirstMixedType{
		Id:       "2",
		Location:  "wecpkbeddsmf",
	})
	db.internalIndex.Index("3", &FirstMixedType{
		Id:       "3",
		Location:  "wecnzm94b80h",
	})
	db.internalIndex.Index("4", &FirstMixedType{
		Id:       "4",
		Location:  "wecpk8tne453",
	})
	db.internalIndex.Index("5", &FirstMixedType{
		Id:       "5",
		Location:  "wecnycjgz1u3",
	})
	db.internalIndex.Index("6", &FirstMixedType{
		Id:       "6",
		Location:  "wecny57t09cu",
	})
	db.internalIndex.Index("7", &FirstMixedType{
		Id:       "7",
		Location:  "wecpkb80s09t",
	})
	db.internalIndex.Index("8", &FirstMixedType{
		Id:       "8",
		Location:  "wecpjbbru3dj",
	})
	db.internalIndex.Index("9", &FirstMixedType{
		Id:       "9",
		Location:  "wecnznn0hzr1",
	})
	db.internalIndex.Index("10", &FirstMixedType{
		Id:       "10",
		Location:  "wecpqgeu2uzw",
	})
	//time.Sleep(5000)
	//
	//if v, err := db.internalIndex.Stats().MarshalJSON(); err==nil{
	//	fmt.Println(string(v))
	//}

	//if err!=nil {
	//	t.Errorf("unexpected error:%v",err)
	//}

	lat, lon := 22.371154, 114.112603
	q := bleve.NewGeoDistanceQuery(lon, lat, "1km")

	req := bleve.NewSearchRequest(q)
	req.Fields = []string{"*"}
	req.SortBy([]string{"_id"})

	sr, err := db.internalIndex.Search(req)
	if err != nil {
		t.Errorf("unexpected error:%v", err)
	}

	if sr.Total == 0 {
		t.Errorf("Size expected: >0, actual %d\n", sr.Total)
	}

	t.Logf("Total result: %d", sr.Total)

	//indexBytes, _ := json.Marshal(db.internalIndex.Mapping())
	//t.Errorf("%s", string(indexBytes))

	//for i := range sr.Hits {
	//	t.Errorf("%s: %s\n", sr.Hits[i].Fields["Name"], sr.Hits[i].Fields["Address"])
	//}

	indexBytes, _ := json.Marshal(db.internalIndex.Mapping())
	t.Logf("%s",string(indexBytes))
	//db.internalIndex.Search()

	// Close the database
	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}*/
