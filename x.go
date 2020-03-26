package dodod

import (
	"errors"
)

var ErrEmptyPath = errors.New("dodod: empty path")
var ErrEmptyPassword = errors.New("dodod: empty password")
var ErrInvalidData = errors.New("dodod: invalid data")
var ErrInvalidDocument = errors.New("dodod: invalid document")

//var ErrDododTypeFieldDoesNotExists = errors.New("dodod: <type> field does not exists")

//var ErrNilPointer = errors.New("dodod: nil pointer")
var ErrDatabasePasswordChangeFailed = errors.New("dodod: database password change failed")
var ErrIndexStorePasswordChangeFailed = errors.New("dodod: index store password change failed")

var ErrDatabaseTransactionFailed = errors.New("dodod: database transaction failed")
var ErrIndexStoreTransactionFailed = errors.New("dodod: index store transaction failed")

var ErrIdCanNotBeEmpty = errors.New("dodod: id can not be empty")
var ErrDatabaseIsNotOpen = errors.New("dodod: database is not open")

// ErrFieldTypeMismatch will occur if the field already registered as different type
var ErrFieldTypeMismatch = errors.New("dodod: field type mismatch")
var ErrDocumentTypeAlreadyRegistered = errors.New("dodod: document type already registered")
var ErrDocumentTypeIsNotRegistered = errors.New("dodod: document type is not registered")

//var ErrInvalidPath = errors.New("dodod: invalid path")

var ErrWrongPassword = errors.New("dodod: wrong password")

//var ErrFailedToOpenDatabase = errors.New("dodod: failed to open database")
//var ErrFailedToCloseDatabase = errors.New("dodod: failed to close database")
var ErrInvalidConfigFile = errors.New("dodod: invalid config file")
var ErrJSONParseFailed = errors.New("dodod: failed to parse json data")

var ErrInvalidBase = errors.New(`dodod: invalid base`)
var ErrInvalidDoc = errors.New(`dodod: invalid doc, nil pointer`)
var ErrInvalidDocNotStruct = errors.New(`dodod: invalid doc, not a struct`)
var ErrUnknownBaseType = errors.New(`dodod: unknown base type`)
var ErrNonBooleanValueForBooleanField = errors.New(`dodod: non-boolean value for boolean field`)
var ErrUnknownMappingField = errors.New(`dodod: tried to set mapping field of unknown type`)

//func IsWrongPassword(err error) bool {
//	return ErrSecretsPassword == err
//}
//
//func IsFailedToOpenDatabase(err error) bool {
//	return ErrFailedToOpenDatabase == err
//}

//IsErrorType checks if the provided value is error or not
func IsErrorType(value interface{}) bool {
	if _, ok := value.(error); ok {
		return ok
	} else {
		return false
	}
}
