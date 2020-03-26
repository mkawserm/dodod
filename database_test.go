package dodod

import (
	"errors"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mkawserm/pasap"
	"os"
	"reflect"
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

	if !reflect.DeepEqual(db.GetRegisteredFields(), []string{"id", "name"}) {
		t.Fatalf("Registered fields are not equal")
	}
}
