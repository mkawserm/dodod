package dodod

import (
	"errors"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mkawserm/pasap"
	"os"
	"sort"
	"testing"
	"time"
)

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
	//		t.Fatalf("error occured while closing, error: %v", err)
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

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []interface{}{&MyTestDocument{Id: "1"}}

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

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []interface{}{&MyTestDocument{Id: "1"}}

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

	err := db.Open()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := []interface{}{&MyTestDocument{Id: "1"}}

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

	if total, _, result, err := db.SimpleSearch("value", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(result) != 5 {
			t.Fatalf("result should be 5 but got %v", len(result))
		}
		if total != 14 {
			t.Fatalf("total result should be 14 but got %v", total)
		}
		doc1 := result[0].(*CustomDocument)
		doc2 := result[1].(*CustomDocument)
		if doc1.Id == doc2.Id {
			t.Fatalf("Document id should not be equal")
		}
		if doc1.CustomField1 == doc2.CustomField1 {
			t.Fatalf("field should not be equal")
		}
		if doc1.CustomField2 == doc2.CustomField2 {
			t.Fatalf("field should not be equal")
		}
		//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	}

	if total, _, result, err := db.SimpleSearch("value", 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(result) != 5 {
			t.Fatalf("result should be 5 but got %v", len(result))
		}
		if total != 14 {
			t.Fatalf("total result should be 14 but got %v", total)
		}
		doc1 := result[0].(*CustomDocument)
		doc2 := result[1].(*CustomDocument)
		if doc1.Id == doc2.Id {
			t.Fatalf("Document id should not be equal")
		}
		if doc1.CustomField1 == doc2.CustomField1 {
			t.Fatalf("field should not be equal")
		}
		if doc1.CustomField2 == doc2.CustomField2 {
			t.Fatalf("field should not be equal")
		}
		//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	}

	if total, _, result, err := db.SimpleSearch("value", 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(result) != 4 {
			t.Fatalf("result should be 4 but got %v", len(result))
		}
		if total != 14 {
			t.Fatalf("total result should be 14 but got %v", total)
		}
		doc1 := result[0].(*CustomDocument)
		doc2 := result[1].(*CustomDocument)
		if doc1.Id == doc2.Id {
			t.Fatalf("Document id should not be equal")
		}
		if doc1.CustomField1 == doc2.CustomField1 {
			t.Fatalf("field should not be equal")
		}
		if doc1.CustomField2 == doc2.CustomField2 {
			t.Fatalf("field should not be equal")
		}
		//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	}

	if total, _, result, err := db.SimpleSearch("value", 15); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(result) != 0 {
			t.Fatalf("result should be 0 but got %v", len(result))
		}
		if total != 14 {
			t.Fatalf("total result should be 14 but got %v", total)
		}
		//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

func TestDatabase_ComplexSearch(t *testing.T) {
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

	sortBy := []string{"-id"}
	queryType := "QueryString"
	limit := 10
	fields := []string{"*"}

	if total, _, result, err := db.ComplexSearch("value", fields, sortBy, queryType, 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(result) != 10 {
			t.Fatalf("result should be 10 but got %v", len(result))
		}
		if total != 14 {
			t.Fatalf("total result should be 14 but got %v", total)
		}

		doc1 := result[0].(*CustomDocument)
		doc2 := result[1].(*CustomDocument)
		if doc1.Id == doc2.Id {
			t.Fatalf("Document id should not be equal")
		}
		if doc1.CustomField1 == doc2.CustomField1 {
			t.Fatalf("field should not be equal")
		}
		if doc1.CustomField2 == doc2.CustomField2 {
			t.Fatalf("field should not be equal")
		}
		//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	}

	if total, _, result, err := db.ComplexSearch("value", fields, sortBy, queryType, 5, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if len(result) != 9 {
			t.Fatalf("result should be 9 but got %v", len(result))
		}
		if total != 14 {
			t.Fatalf("total result should be 14 but got %v", total)
		}
		//fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

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

func (m *FirstMixedType) Type() string {
	return "FirstMixedType"
}

func (m *FirstMixedType) GetId() string {
	return m.Id
}

type SecondMixedType struct {
	Id        string `json:"id"`
	MixedType string `json:"mixed_type"`

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
			Field11:   "SMTF 11",
			Field12:   "SMTF 12",
			Field13:   213,
			Field14:   214.2,
			Field15:   false,
			Field16:   time.Now(),
			Field17:   []byte{21, 22, 23, 24, 25},
			Field18:   []string{"21", "22", "23"},
			Field19:   []int64{211111111111111, 311111111111122, 411111111111133},
			Field20:   []float64{2111111111111.11, 3111111111111.22, 4111111111111.33},
		},
		&ThirdMixedType{
			Id:        "3",
			MixedType: "ThirdMixedType",
			Field21:   "TMTF 21",
			Field22:   "TMTF 22",
			Field23:   323,
			Field24:   324.3,
			Field25:   true,
			Field26:   time.Now(),
			Field27:   []byte{31, 32, 33, 34, 35},
			Field28:   []string{"31", "32", "33"},
			Field29:   []int64{311111111111111, 411111111111122, 511111111111133},
			Field30:   []float64{3111111111111.11, 4111111111111.22, 5111111111111.33},
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
	db.SetDbPath(dbPath)
	db.SetupDefaults()

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

	// Find Id
	if total, _, result, err := db.FindId("ThirdMixedType", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if total != 1 {
			t.Fatalf("Total result should be 1")
		}

		if result[0].Id != "3" {
			t.Fatalf("Id be 3")
		}
	}

	// Facet search
	if _, data, err := db.FacetSearch([]map[string]interface{}{{"facetName": "Types", "queryInput": "mixed_type", "facetLimit": 10}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		//t.Errorf("%v", data)
		//t.Errorf("%v", len(data["Types"]))
		if len(data["Types"]) != 4 {
			t.Fatalf("Facet types should be 4")
		}
	}

	// Facet search
	if _, data, err := db.FacetSearch([]map[string]interface{}{{"facetName": "Types", "queryInput": "mixed_type", "facetLimit": 1}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		//t.Errorf("%v", data)
		//t.Errorf("%v", len(data["Types"]))
		if len(data["Types"]) != 3 {
			t.Fatalf("Facet types should be 3")
		}
	}

	// BleveComplex SimpleSearch
	sortBy := []string{"-id"}
	queryType := "QueryString"
	limit := 10
	fields := []string{"*"}
	if result, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, queryType, 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if result.Total != 1 {
			t.Fatalf("result should be one")
		}
	}

	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "FuzzyQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "MatchAllQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "MatchQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "MatchPhraseQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "RegexpQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "TermQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "WildcardQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "PrefixQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := db.BleveComplexSearch("ThirdMixedType", fields, sortBy, "", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	//Complex SimpleSearch
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "QueryString", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "FuzzyQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "MatchAllQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "MatchQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "MatchPhraseQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "RegexpQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "TermQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "WildcardQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "PrefixQuery", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, _, err := db.ComplexSearch("ThirdMixedType", fields, sortBy, "", 0, limit); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Search
	input := make(map[string]interface{})
	if data, err := db.Search(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if data == nil {
			t.Fatalf("data should not be nil")
		}
	}

	// Close the database
	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}
