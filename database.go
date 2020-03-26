package dodod

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	"github.com/mkawserm/bdodb"
	"github.com/mkawserm/pasap"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

type BleveIndexOpener struct {
}

func (b *BleveIndexOpener) BleveIndex(dbPath string,
	indexMapping *mapping.IndexMappingImpl,
	indexName string,
	config map[string]interface{}) (bleve.Index, error) {

	return bdodb.BleveIndex(dbPath, indexMapping, indexName, config)
}

type Database struct {
	passwordHasher        pasap.PasswordHasher
	encoderCredentialsRW  pasap.EncoderCredentialsRW
	verifierCredentialsRW pasap.VerifierCredentialsRW
	indexOpener           IndexOpener

	indexMapping *mapping.IndexMappingImpl

	dbPath     string
	dbPassword string

	secretKey           []byte
	encodedKey          string
	isPasswordProtected bool
	isDbReady           bool

	fieldsRegistryCache   map[string]string
	documentRegistryCache map[string]Document

	internalSearchResultLimit int
	internalIndex             bleve.Index
	internalDb                *badger.DB
	internalIndexStoreName    string
}

func (db *Database) initAll() {
	db.internalSearchResultLimit = 20

	if db.internalIndexStoreName == "" {
		db.internalIndexStoreName = "badger"
	}

	if db.fieldsRegistryCache == nil {
		db.fieldsRegistryCache = make(map[string]string)
	}

	if db.documentRegistryCache == nil {
		db.documentRegistryCache = make(map[string]Document)
	}
}

func (db *Database) SetDbPath(dbPath string) {
	db.dbPath = dbPath
}

func (db *Database) SetDbPassword(dbPassword string) {
	db.dbPassword = dbPassword
}

func (db *Database) Setup(passwordHasher pasap.PasswordHasher,
	encoderCredentialsRW pasap.EncoderCredentialsRW,
	verifierCredentialsRW pasap.VerifierCredentialsRW,
	indexOpener IndexOpener) {

	db.passwordHasher = passwordHasher
	db.encoderCredentialsRW = encoderCredentialsRW
	db.verifierCredentialsRW = verifierCredentialsRW
	db.indexOpener = indexOpener

	db.initAll()
	db.initIndexMapping()
}

func (db *Database) SetSearchResultLimit(n int) {
	db.internalSearchResultLimit = n
}

func (db *Database) SetPasswordHasher(passwordHasher pasap.PasswordHasher) {
	db.passwordHasher = passwordHasher
}

func (db *Database) SetEncoderCredentialsRW(encoderCredentialsRW pasap.EncoderCredentialsRW) {
	db.encoderCredentialsRW = encoderCredentialsRW
}

func (db *Database) SetVerifierCredentialsRW(verifierCredentialsRW pasap.VerifierCredentialsRW) {
	db.verifierCredentialsRW = verifierCredentialsRW
}

func (db *Database) SetIndexOpener(opener IndexOpener) {
	db.indexOpener = opener
}

func (db *Database) SetIndexStoreName(indexStoreName string) {
	db.internalIndexStoreName = indexStoreName
}

//func (db *Database) SetIndexMapping(indexMapping *mapping.IndexMappingImpl) {
//	db.indexMapping = indexMapping
//}

//func (db *Database) GetIndexMapping() *mapping.IndexMappingImpl {
//	return db.indexMapping
//}

func (db *Database) SetupDefaults() {
	db.passwordHasher = pasap.NewArgon2idHasher()
	db.encoderCredentialsRW = &pasap.ByteBasedEncoderCredentials{}
	db.verifierCredentialsRW = &pasap.ByteBasedVerifierCredentials{}
	db.indexOpener = &BleveIndexOpener{}

	db.initAll()
	db.initIndexMapping()
}

func (db *Database) Open() error {
	db.initAll()
	db.initIndexMapping()

	if db.dbPath == "" {
		return ErrEmptyPath
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

func (db *Database) Close() error {
	db.isDbReady = false

	var err1 error
	var err2 error

	if db.internalIndex != nil {
		err1 = db.internalIndex.Close()
	}
	if db.internalDb != nil {
		err2 = db.internalDb.Close()
	}

	if err1 == nil && err2 == nil {
		return nil
	} else if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}

	return nil
}

func (db *Database) initIndexMapping() {
	if db.indexMapping == nil {
		db.indexMapping = bleve.NewIndexMapping()
	}
}

func (db *Database) RegisterDocument(d interface{}) error {
	db.initAll()
	db.initIndexMapping()
	var document Document

	if n, ok := d.(Document); !ok {
		return ErrInvalidDocument
	} else {
		document = n
	}

	if _, exists := db.documentRegistryCache[document.Type()]; exists {
		return ErrDocumentTypeAlreadyRegistered
	}

	canRegister := true
	fields := ExtractFields(document)
	for k, v := range fields {
		f, exists := db.fieldsRegistryCache[k]
		if exists {
			if f != v {
				canRegister = false
				break
			}
		}
	}

	if !canRegister {
		return ErrFieldTypeMismatch
	}
	if err := registerDocumentMapping(db.indexMapping, document); err != nil {
		return err
	}

	for k, v := range fields {
		_, exists := db.fieldsRegistryCache[k]
		if !exists {
			db.fieldsRegistryCache[k] = v
		}
	}

	db.documentRegistryCache[document.Type()] = document

	return nil
}

func (db *Database) GetRegisteredFields() []string {
	keys := make([]string, 0, len(db.fieldsRegistryCache))
	for k := range db.fieldsRegistryCache {
		keys = append(keys, k)
	}
	return keys
}

func (db *Database) IsDatabaseReady() bool {
	return db.isDbReady
}

func (db *Database) EncodeDocument(document interface{}) ([]byte, error) {
	var err error
	var jsonData []byte

	var data Document
	if n, ok := document.(Document); !ok {
		return nil, ErrInvalidDocument
	} else {
		data = n
	}

	documentType := []byte(data.Type())
	jsonData, err = json.Marshal(data)

	if err != nil {
		return nil, err
	}

	typeLength := uint32(len(documentType))
	dataLength := uint32(len(jsonData))

	typeLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(typeLengthBytes, typeLength)

	dataLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLengthBytes, dataLength)

	var output bytes.Buffer

	output.Write(typeLengthBytes)
	output.Write(documentType)
	output.Write(dataLengthBytes)
	output.Write(jsonData)

	return output.Bytes(), nil
}

func (db *Database) DecodeDocument(data []byte) (interface{}, error) {
	//var err error
	var doc interface{}

	if len(data) < 8 {
		return nil, ErrInvalidData
	}

	var documentTypeLength uint32
	//var jsonDataLength uint32

	documentTypeLength = binary.BigEndian.Uint32(data[0:4])
	if len(data) < (8 + int(documentTypeLength)) {
		return nil, ErrInvalidData
	}

	//jsonDataLength = binary.BigEndian.Uint32(data[4+documentTypeLength:4+documentTypeLength+4])
	documentType := string(data[4 : 4+documentTypeLength])
	if v, exists := db.documentRegistryCache[documentType]; !exists {
		return nil, ErrDocumentTypeIsNotRegistered
	} else {
		indirect := reflect.Indirect(reflect.ValueOf(v))
		newIndirect := reflect.New(indirect.Type())
		doc = newIndirect.Interface()
	}

	jsonData := data[4+documentTypeLength+4:]
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (db *Database) DecodeDocumentUsingInterface(data []byte, document interface{}) error {
	if len(data) < 8 {
		return ErrInvalidData
	}

	var documentTypeLength uint32
	//var jsonDataLength uint32

	documentTypeLength = binary.BigEndian.Uint32(data[0:4])

	if len(data) < (8 + int(documentTypeLength)) {
		return ErrInvalidData
	}

	jsonData := data[4+documentTypeLength+4:]
	if err := json.Unmarshal(jsonData, document); err != nil {
		return err
	}

	return nil
}

func (db *Database) Create(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	var err1 error
	var err2 error

	internalBatchTxn := db.internalDb.NewTransaction(true)
	defer internalBatchTxn.Discard()

	batch := db.internalIndex.NewBatch()
	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}

		if id == "" {
			return ErrIdCanNotBeEmpty
		}

		if jsonData, err := db.EncodeDocument(d); err == nil {
			if err := internalBatchTxn.Set([]byte(id), jsonData); err != nil {
				return err
			}
		} else {
			return err
		}

		if err := batch.Index(id, d); err != nil {
			return err
		}
	}

	err1 = internalBatchTxn.Commit()
	if err1 != nil {
		return ErrDatabaseTransactionFailed
	}

	err2 = db.internalIndex.Batch(batch)
	if err2 != nil {
		return ErrIndexStoreTransactionFailed
	}

	return nil
}

func (db *Database) Update(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	var err1 error
	var err2 error

	internalBatchTxn := db.internalDb.NewTransaction(true)
	defer internalBatchTxn.Discard()

	batch := db.internalIndex.NewBatch()
	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}

		if id == "" {
			return ErrIdCanNotBeEmpty
		}

		if err := internalBatchTxn.Delete([]byte(id)); err != nil {
			return err
		}

		if jsonData, err := db.EncodeDocument(d); err == nil {
			if err := internalBatchTxn.Set([]byte(id), jsonData); err != nil {
				return err
			}
		} else {
			return err
		}

		// batch.Delete(id)
		if err := batch.Index(id, d); err != nil {
			return err
		}
	}

	err1 = internalBatchTxn.Commit()
	if err1 != nil {
		return ErrDatabaseTransactionFailed
	}

	err2 = db.internalIndex.Batch(batch)
	if err2 != nil {
		return ErrIndexStoreTransactionFailed
	}

	return nil
}

func (db *Database) Delete(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	var err1 error
	var err2 error

	internalBatchTxn := db.internalDb.NewTransaction(true)
	defer internalBatchTxn.Discard()

	batch := db.internalIndex.NewBatch()
	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}

		if id == "" {
			return ErrIdCanNotBeEmpty
		}

		if err := internalBatchTxn.Delete([]byte(id)); err != nil {
			return err
		}

		batch.Delete(id)
	}

	err1 = internalBatchTxn.Commit()
	if err1 != nil {
		return ErrDatabaseTransactionFailed
	}

	err2 = db.internalIndex.Batch(batch)
	if err2 != nil {
		return ErrIndexStoreTransactionFailed
	}

	return nil
}

func (db *Database) AddDocument(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	var err1 error

	internalBatchTxn := db.internalDb.NewTransaction(true)
	defer internalBatchTxn.Discard()

	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}

		if id == "" {
			return ErrIdCanNotBeEmpty
		}

		if jsonData, err := db.EncodeDocument(d); err == nil {
			if err := internalBatchTxn.Set([]byte(id), jsonData); err != nil {
				return err
			}
		} else {
			return err
		}

	}

	err1 = internalBatchTxn.Commit()
	if err1 != nil {
		return ErrDatabaseTransactionFailed
	}

	return nil
}

func (db *Database) UpdateDocument(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	var err1 error

	internalBatchTxn := db.internalDb.NewTransaction(true)
	defer internalBatchTxn.Discard()

	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}

		if id == "" {
			return ErrIdCanNotBeEmpty
		}

		if err := internalBatchTxn.Delete([]byte(id)); err != nil {
			return err
		}

		if jsonData, err := db.EncodeDocument(d); err == nil {
			if err := internalBatchTxn.Set([]byte(id), jsonData); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err1 = internalBatchTxn.Commit()
	if err1 != nil {
		return ErrDatabaseTransactionFailed
	}

	return nil
}

func (db *Database) DeleteDocument(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	var err1 error

	internalBatchTxn := db.internalDb.NewTransaction(true)
	defer internalBatchTxn.Discard()

	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}

		if id == "" {
			return ErrIdCanNotBeEmpty
		}

		if err := internalBatchTxn.Delete([]byte(id)); err != nil {
			return err
		}
	}

	err1 = internalBatchTxn.Commit()
	if err1 != nil {
		return ErrDatabaseTransactionFailed
	}

	return nil
}

func (db *Database) Read(data []interface{}) (uint64, error) {
	if !db.IsDatabaseReady() {
		return 0, ErrDatabaseIsNotOpen
	}

	internalBatchTxn := db.internalDb.NewTransaction(false)
	defer internalBatchTxn.Discard()

	var readCount uint64 = 0
	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		}

		if id == "" {
			continue
		}

		if item, err := internalBatchTxn.Get([]byte(id)); err == nil {
			if value, err := item.ValueCopy(nil); err == nil {
				if err := db.DecodeDocumentUsingInterface(value, d); err == nil {
					readCount = readCount + 1
				} else {
					return readCount, err
				}
			} else {
				return readCount, err
			}
		} else {
			if err == badger.ErrKeyNotFound {
				continue
			} else {
				return readCount, err
			}
		}
	}

	return readCount, nil
}

func (db *Database) ReadUsingId(data []string) (uint64, []interface{}, error) {
	if !db.IsDatabaseReady() {
		return 0, nil, ErrDatabaseIsNotOpen
	}

	internalBatchTxn := db.internalDb.NewTransaction(false)
	defer internalBatchTxn.Discard()

	output := make([]interface{}, len(data), len(data))
	var readCount = 0
	for _, id := range data {
		if id == "" {
			continue
		}

		if item, err := internalBatchTxn.Get([]byte(id)); err == nil {
			if value, err := item.ValueCopy(nil); err == nil {
				if doc, err := db.DecodeDocument(value); err == nil {
					output[readCount] = doc
					readCount = readCount + 1
				}
			}
		}
	}
	return uint64(readCount), output[:readCount], nil
}

func (db *Database) ReadUsingIdWithError(data []string) (uint64, []interface{}, error) {
	if !db.IsDatabaseReady() {
		return 0, nil, ErrDatabaseIsNotOpen
	}

	internalBatchTxn := db.internalDb.NewTransaction(false)
	defer internalBatchTxn.Discard()

	output := make([]interface{}, len(data), len(data))
	var readCount uint64 = 0
	for i, id := range data {
		if id == "" {
			continue
		}

		if item, err := internalBatchTxn.Get([]byte(id)); err == nil {
			if value, err := item.ValueCopy(nil); err == nil {
				if doc, err := db.DecodeDocument(value); err == nil {
					output[i] = doc
					readCount = readCount + 1
				} else {
					output[i] = err
				}
			} else {
				output[i] = err
			}
		} else {
			output[i] = err
		}
	}

	return readCount, output, nil
}

func (db *Database) CreateIndex(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	batch := db.internalIndex.NewBatch()
	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}
		if id == "" {
			return ErrIdCanNotBeEmpty
		}

		if err := batch.Index(id, d); err != nil {
			return err
		}
	}

	return db.internalIndex.Batch(batch)
}

func (db *Database) UpdateIndex(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	batch := db.internalIndex.NewBatch()
	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}
		if id == "" {
			return ErrIdCanNotBeEmpty
		}
		batch.Delete(id)
		if err := batch.Index(id, d); err != nil {
			return err
		}
	}

	return db.internalIndex.Batch(batch)
}

func (db *Database) DeleteIndex(data []interface{}) error {
	if !db.IsDatabaseReady() {
		return ErrDatabaseIsNotOpen
	}

	batch := db.internalIndex.NewBatch()
	for _, d := range data {
		var id string
		if n, ok := d.(Document); ok {
			id = n.GetId()
		} else {
			return ErrInvalidDocument
		}
		if id == "" {
			return ErrIdCanNotBeEmpty
		}
		batch.Delete(id)
	}

	return db.internalIndex.Batch(batch)
}

func (db *Database) Search(queryInput string, offset int) (total uint64,
	queryTime time.Duration,
	result []interface{},
	err error) {

	q := bleve.NewQueryStringQuery(queryInput)
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.From = offset
	searchRequest.Size = db.internalSearchResultLimit
	//searchRequest.Fields = db.GetRegisteredFields()

	var searchResult *bleve.SearchResult
	searchResult, err = db.internalIndex.Search(searchRequest)
	if err != nil {
		return
	}

	queryTime = searchResult.Took
	total = searchResult.Total

	idList := make([]string, len(searchResult.Hits), len(searchResult.Hits))

	for i, v := range searchResult.Hits {
		idList[i] = v.ID
	}

	start := time.Now()
	if _, data, err := db.ReadUsingId(idList); err != nil {
		return 0, 0, nil, err
	} else {
		result = data
	}
	elapsed := time.Now().Sub(start)

	return total, queryTime + elapsed, result, err
}

func (db *Database) FacetSearch(facetInput []FaceInput) (queryTime time.Duration,
	data map[string][]FacetOutput,
	err error) {

	q := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Size = 0
	for _, fi := range facetInput {
		searchRequest.AddFacet(fi.FacetName, bleve.NewFacetRequest(fi.QueryInput, fi.FacetLimit))
	}

	var searchResult *bleve.SearchResult
	searchResult, err = db.internalIndex.Search(searchRequest)
	if err != nil {
		return
	}

	queryTime = searchResult.Took

	if len(searchResult.Facets) > 0 {
		data = make(map[string][]FacetOutput)
		for fn, f := range searchResult.Facets {
			data[fn] = make([]FacetOutput, 0, len(f.Terms)+1)
			data[fn] = append(data[fn], FacetOutput{TermName: fn, TermCount: f.Total})

			for _, t := range f.Terms {
				fo := FacetOutput{
					TermName:  t.Term,
					TermCount: t.Count,
				}
				data[fn] = append(data[fn], fo)
			}

			if f.Other != 0 {
				fo := FacetOutput{
					TermName:  "Extras",
					TermCount: f.Other,
				}
				data[fn] = append(data[fn], fo)
			}
		}
	}

	return
}

func (db *Database) ComplexSearch(queryInput string, sortBy []string, queryType string, offset int, limit int) (
	total uint64,
	queryTime time.Duration,
	result []interface{},
	err error) {

	var q query.Query

	if queryType == "QueryString" {
		q = bleve.NewQueryStringQuery(queryInput)
	} else if queryType == "FuzzyQuery" {
		q = bleve.NewFuzzyQuery(queryInput)
	} else if queryType == "MatchAllQuery" {
		q = bleve.NewMatchAllQuery()
	} else if queryType == "MatchQuery" {
		q = bleve.NewMatchQuery(queryInput)
	} else if queryType == "MatchPhraseQuery" {
		q = bleve.NewMatchPhraseQuery(queryInput)
	} else if queryType == "RegexpQuery" {
		q = bleve.NewRegexpQuery(queryInput)
	} else if queryType == "TermQuery" {
		q = bleve.NewTermQuery(queryInput)
	} else if queryType == "WildcardQuery" {
		q = bleve.NewWildcardQuery(queryInput)
	} else if queryType == "PrefixQuery" {
		q = bleve.NewPrefixQuery(queryInput)
	} else {
		q = bleve.NewQueryStringQuery(queryInput)
	}
	//else if queryType == "PhraseQuery" {
	//	q = bleve.NewPhraseQuery(queryInput)
	//}

	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.From = offset
	searchRequest.Size = limit
	searchRequest.SortBy(sortBy)
	//searchRequest.Fields = db.GetRegisteredFields()

	var searchResult *bleve.SearchResult
	searchResult, err = db.internalIndex.Search(searchRequest)
	if err != nil {
		return
	}

	queryTime = searchResult.Took
	total = searchResult.Total

	idList := make([]string, len(searchResult.Hits), len(searchResult.Hits))

	for i, v := range searchResult.Hits {
		idList[i] = v.ID
	}

	start := time.Now()
	if _, data, err := db.ReadUsingId(idList); err != nil {
		return 0, 0, nil, err
	} else {
		result = data
	}
	elapsed := time.Now().Sub(start)

	return total, queryTime + elapsed, result, err
}

func (db *Database) BleveSearch(req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	if !db.IsDatabaseReady() {
		return nil, ErrDatabaseIsNotOpen
	}

	return db.internalIndex.Search(req)
}

func (db *Database) BleveSearchInContext(ctx context.Context, req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	if !db.IsDatabaseReady() {
		return nil, ErrDatabaseIsNotOpen
	}

	return db.internalIndex.SearchInContext(ctx, req)
}

func (db *Database) isDbExists() bool {
	fi, err := os.Stat(db.dbPath + "/dodod.json")
	return err == nil && fi != nil
}

func (db *Database) readConfig() (bool, error) {
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

	if val, ok := jsonMap["indexStoreName"].(string); ok {
		db.internalIndexStoreName = val
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

func (db *Database) isPasswordValid() (bool, error) {
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

func (db *Database) writeConfig() (bool, error) {
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
	jsonMap["indexStoreName"] = db.internalIndexStoreName

	data, err := json.Marshal(jsonMap)
	if err != nil {
		return false, err
	}

	if err := ioutil.WriteFile(db.dbPath+"/dodod.json", data, 0700); err != nil {
		return false, err
	}

	return true, nil
}

func (db *Database) ensurePath() {
	//if _, err := os.Stat(db.dbPath); err != nil {
	//	_ = os.MkdirAll(db.dbPath, os.FileMode(0700))
	//}

	if _, err := os.Stat(db.dbPath + "/database"); err != nil {
		_ = os.MkdirAll(db.dbPath+"/database", os.FileMode(0700))
	}
}

func (db *Database) openDb() error {
	var index bleve.Index
	var err error

	if db.internalIndexStoreName == "badger" {
		index, err = db.indexOpener.BleveIndex(db.dbPath,
			db.indexMapping,
			upsidedown.Name,
			map[string]interface{}{
				"BdodbConfig": &bdodb.Config{
					EncryptionKey: db.secretKey,
					Logger:        DefaultLogger,
				},
			})
	} else {
		index, err = db.indexOpener.BleveIndex(db.dbPath,
			db.indexMapping,
			scorch.Name,
			map[string]interface{}{
				"BdodbConfig": &bdodb.Config{
					EncryptionKey: db.secretKey,
				},
			})
	}

	if err != nil {
		return err
	}

	/* Main DataStore */
	opt := badger.DefaultOptions(db.dbPath + "/database")
	opt.ReadOnly = false
	opt.Truncate = true
	opt.TableLoadingMode = options.LoadToRAM
	opt.ValueLogLoadingMode = options.MemoryMap
	opt.Compression = options.Snappy
	opt.EncryptionKey = db.secretKey
	opt.Logger = DefaultLogger

	if badgerDb, err := badger.Open(opt); err != nil {
		_ = index.Close()
		return err
	} else {
		db.internalDb = badgerDb
	}

	db.internalIndex = index
	db.isDbReady = true

	return nil
}

func (db *Database) ChangePassword(newPassword string) error {
	if db.dbPath == "" {
		return ErrEmptyPath
	}

	if db.dbPassword == "" {
		return ErrEmptyPassword
	}

	if ok, err := db.readConfig(); !ok {
		return err
	}

	if ok, err := db.isPasswordValid(); !ok {
		return err
	}

	oldKey := db.secretKey

	if err := copyFile(db.dbPath+"/dodod.json", db.dbPath+"/database.dodod.json.backup"); err != nil {
		return err
	}

	if db.internalIndexStoreName == "badger" {
		if err := copyFile(db.dbPath+"/dodod.json", db.dbPath+"/indexstore.dodod.json.backup"); err != nil {
			return err
		}
	}

	db.dbPassword = newPassword
	if ok, err := db.writeConfig(); !ok {
		return err
	}
	newKey := db.secretKey

	opt := badger.KeyRegistryOptions{
		Dir:                           db.dbPath + "/database",
		ReadOnly:                      true,
		EncryptionKey:                 oldKey,
		EncryptionKeyRotationDuration: 10 * 24 * time.Hour,
	}

	kr, err := badger.OpenKeyRegistry(opt)
	if err != nil {
		return err
	}
	opt.EncryptionKey = newKey
	err = badger.WriteKeyRegistry(kr, opt)
	if err != nil {
		DefaultLogger.Errorf("%v", err)
		return ErrDatabasePasswordChangeFailed
	}

	if db.internalIndexStoreName == "badger" {
		// index store
		opt1 := badger.KeyRegistryOptions{
			Dir:                           db.dbPath + "/store",
			ReadOnly:                      true,
			EncryptionKey:                 oldKey,
			EncryptionKeyRotationDuration: 10 * 24 * time.Hour,
		}

		kr1, err1 := badger.OpenKeyRegistry(opt1)
		if err1 != nil {
			return err1
		}
		opt1.EncryptionKey = newKey
		err = badger.WriteKeyRegistry(kr1, opt1)
		if err != nil {
			DefaultLogger.Errorf("%v", err)
			return ErrIndexStorePasswordChangeFailed
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
