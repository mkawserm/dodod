package dodod

import "errors"

//var ErrInvalidPath = errors.New("dodod: invalid path")

var ErrWrongPassword = errors.New("dodod: wrong password")

//var ErrFailedToOpenDatabase = errors.New("dodod: failed to open database")
//var ErrFailedToCloseDatabase = errors.New("dodod: failed to close database")
var ErrInvalidConfigFile = errors.New("dodod: invalid config file")
var ErrJSONParseFailed = errors.New("dodod: failed to parse json data")

//func IsWrongPassword(err error) bool {
//	return ErrSecretsPassword == err
//}
//
//func IsFailedToOpenDatabase(err error) bool {
//	return ErrFailedToOpenDatabase == err
//}
