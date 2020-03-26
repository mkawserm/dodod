package dodod

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mkawserm/pasap"
	"os"
	"sort"
	"testing"
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
	dbPassword := "password"
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
	dbPassword := "password"
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
	dbPassword := "password"

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

	if _, err := db.Read([]interface{}{&MyTestDocument{
		Id:   "1",
		Name: "Test1",
	}}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, _, err := db.ReadUsingId([]string{"1"}); err != ErrDatabaseIsNotOpen {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := db.ReadUsingIdWithError([]string{"1"}); err != ErrDatabaseIsNotOpen {
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

	if n, err := db.Read(data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
	}

	if n, data, err := db.ReadUsingId([]string{"1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
		if len(data) != 1 {
			t.Fatalf("read failure")
		}
	}

	if n, data, err := db.ReadUsingId([]string{"1", "2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		if n != 1 {
			t.Fatalf("read failure")
		}
		if len(data) != 1 {
			t.Fatalf("read failure")
		}
	}

	if n, data, err := db.ReadUsingIdWithError([]string{"1", "2"}); err != nil {
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

	if n, err := db.Read(data); err != nil {
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

	if n, err := db.Read(data); err != nil {
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

func TestDatabase_UpdateIndex(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := "password"

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
	dbPassword := "password"

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
	dbPassword := "password"

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

	total, queryTime, result, err := db.Search("value", 0)

	fmt.Println("Total: ", total, "| Query time:", queryTime, "| Result: ", result, "| Err:", err)

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}

func TestDatabase_EncodeDecodeDocument(t *testing.T) {
	t.Helper()

	dbPath := "/tmp/dodod"
	dbPassword := "password"

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
		if n, err := db.DecodeDocument(d); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else {
			fmt.Println(n.(*CustomDocument))
		}
	}

	if err := db.Close(); err != nil {
		t.Fatalf("error occured while closing, error: %v", err)
	}
}
