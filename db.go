package dodod

import (
	"encoding/json"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mkawserm/bdodb"
	"github.com/mkawserm/pasap"
	"io/ioutil"
	"os"
)

type BleveIndexOpener struct {
}

func (b *BleveIndexOpener) BleveIndex(dbPath string,
	indexMappingImpl *mapping.IndexMappingImpl,
	indexName string,
	config map[string]interface{}) (bleve.Index, error) {
	return bdodb.BleveIndex(dbPath, indexMappingImpl, indexName, config)
}

type DbCredentialsBasic struct {
	Path     string
	Password string
}

func (d *DbCredentialsBasic) ReadPath() (dbPath string, err error) {
	return d.Path, nil
}

func (d *DbCredentialsBasic) ReadPassword() (password string, err error) {
	return d.Password, nil
}

type Db struct {
	dbCredentials         DbCredentials
	passwordHasher        pasap.PasswordHasher
	encoderCredentialsRW  pasap.EncoderCredentialsRW
	verifierCredentialsRW pasap.VerifierCredentialsRW
	indexOpener           IndexOpener

	indexMappingImpl *mapping.IndexMappingImpl
	index            bleve.Index

	dbPath     string
	dbPassword string

	secretKey           []byte
	encodedKey          string
	isPasswordProtected bool
}

func (db *Db) Setup(dbCredentials DbCredentials,
	passwordHasher pasap.PasswordHasher,
	encoderCredentialsRW pasap.EncoderCredentialsRW,
	verifierCredentialsRW pasap.VerifierCredentialsRW,
	indexOpener IndexOpener,
	indexMappingImpl *mapping.IndexMappingImpl) {

	db.dbCredentials = dbCredentials
	db.passwordHasher = passwordHasher
	db.encoderCredentialsRW = encoderCredentialsRW
	db.verifierCredentialsRW = verifierCredentialsRW
	db.indexOpener = indexOpener
	db.indexMappingImpl = indexMappingImpl
}

func (db *Db) SetDbCredentials(credentials DbCredentials) {
	db.dbCredentials = credentials
}

func (db *Db) SetPasswordHasher(passwordHasher pasap.PasswordHasher) {
	db.passwordHasher = passwordHasher
}

func (db *Db) SetEncoderCredentialsRW(encoderCredentialsRW pasap.EncoderCredentialsRW) {
	db.encoderCredentialsRW = encoderCredentialsRW
}

func (db *Db) SetVerifierCredentialsRW(verifierCredentialsRW pasap.VerifierCredentialsRW) {
	db.verifierCredentialsRW = verifierCredentialsRW
}

func (db *Db) SetIndexOpener(opener IndexOpener) {
	db.indexOpener = opener
}

func (db *Db) SetIndexMappingImpl(indexMappingImpl *mapping.IndexMappingImpl) {
	db.indexMappingImpl = indexMappingImpl
}

func (db *Db) SetupDefaults() {
	db.passwordHasher = pasap.NewArgon2idHasher()
	db.encoderCredentialsRW = &pasap.ByteBasedEncoderCredentials{}
	db.verifierCredentialsRW = &pasap.ByteBasedVerifierCredentials{}
	db.indexMappingImpl = bleve.NewIndexMapping()
	db.indexOpener = &BleveIndexOpener{}
}

func (db *Db) Open() error {
	if path, err := db.dbCredentials.ReadPath(); err != nil {
		return err
	} else {
		db.dbPath = path
	}

	if password, err := db.dbCredentials.ReadPassword(); err != nil {
		return err
	} else {
		db.dbPassword = password
	}

	// Ensure path exists or create path
	db.ensurePath()
	exists := db.isDbExists()

	//fmt.Println("Exists: ", exists)

	if exists {
		if _, readError := db.readConfig(); readError != nil {
			return readError
		}
	} else {
		if _, writeError := db.writeConfig(); writeError != nil {
			return writeError
		}
	}
	return db.openDb()
}

func (db *Db) Close() error {
	if db.index != nil {
		return db.index.Close()
	}
	return nil
}

func (db *Db) isDbExists() bool {
	fi, err := os.Stat(db.dbPath + "/dodod.json")
	return err == nil && fi != nil
}

func (db *Db) readConfig() (bool, error) {
	data, _ := ioutil.ReadFile(db.dbPath + "/dodod.json")
	if len(data) == 0 {
		return false, ErrInvalidConfigFile
	}

	jsonMap := make(map[string]interface{})
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return false, ErrJSONParseFailed
	}

	if val, ok := jsonMap["encodedKey"].(string); ok {
		db.encodedKey = val
	}

	if val, ok := jsonMap["isPasswordProtected"].(bool); ok {
		db.isPasswordProtected = val
	}

	if db.isPasswordProtected {
		v, e := db.isPasswordValid()
		if !v {
			return false, e
		}
	}

	return true, nil
}

func (db *Db) isPasswordValid() (bool, error) {
	if err := db.verifierCredentialsRW.SetPassword([]byte(db.dbPassword)); err != nil {
		return false, err
	}

	if err := db.verifierCredentialsRW.SetEncodedKey([]byte(db.encodedKey)); err != nil {
		return false, err
	}

	secretKey, ok, err := db.passwordHasher.Verify(db.verifierCredentialsRW)

	if err != nil {
		return false, err
	}

	if ok {
		db.secretKey = secretKey
		return true, nil
	} else {
		return false, ErrWrongPassword
	}
}

func (db *Db) writeConfig() (bool, error) {
	db.isPasswordProtected = false
	db.encodedKey = ""

	if len(db.dbPassword) != 0 {
		if err := db.encoderCredentialsRW.SetSalt(pasap.GetSalt(16, nil)); err != nil {
			return false, err
		}
		if err := db.encoderCredentialsRW.SetPassword([]byte(db.dbPassword)); err != nil {
			return false, err
		}

		secretKey, encodedKey, err := db.passwordHasher.Encode(db.encoderCredentialsRW)

		if err != nil {
			return false, err
		}

		db.encodedKey = string(encodedKey)
		db.secretKey = secretKey
		db.isPasswordProtected = true
	}

	jsonMap := make(map[string]interface{})
	jsonMap["encodedKey"] = db.encodedKey
	jsonMap["isPasswordProtected"] = db.isPasswordProtected

	data, err := json.Marshal(jsonMap)
	if err != nil {
		return false, err
	}

	if err := ioutil.WriteFile(db.dbPath+"/dodod.json", data, 0700); err != nil {
		return false, err
	}

	return true, nil
}

func (db *Db) ensurePath() {
	if _, err := os.Stat(db.dbPath); err != nil {
		_ = os.MkdirAll(db.dbPath, os.FileMode(0700))
	}
}

func (db *Db) openDb() error {
	index, err := db.indexOpener.BleveIndex(db.dbPath,
		db.indexMappingImpl,
		upsidedown.Name,
		map[string]interface{}{
			"BdodbConfig": &bdodb.Config{
				EncryptionKey: db.secretKey,
			},
		})

	if err != nil {
		return err
	}
	db.index = index

	return nil
}
