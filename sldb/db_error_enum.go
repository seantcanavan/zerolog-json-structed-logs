package sldb

import (
	"net/http"
)

// EnumDBErrorType is a string type for representing database error constants.
type EnumDBErrorType string

// Enumeration of common database errors as string type constants.
const (
	ErrDBAccessDenied       EnumDBErrorType = "Access Denied"
	ErrDBConnectionFailed   EnumDBErrorType = "Connection Failed"
	ErrDBConstraintViolated EnumDBErrorType = "Constraint Violation"
	ErrDBDataOutOfRange     EnumDBErrorType = "Data Out of Range"
	ErrDBDuplicateEntry     EnumDBErrorType = "Duplicate Entry"
	ErrDBForeignKeyViolated EnumDBErrorType = "Foreign Key Violation"
	ErrDBInvalidTransaction EnumDBErrorType = "Invalid Transaction"
	ErrDBQueryInterrupted   EnumDBErrorType = "Query Interrupted"
	ErrDBRecordNotFound     EnumDBErrorType = "Record Not Found"
	ErrDBSyntaxError        EnumDBErrorType = "Syntax Error"
	ErrDBTimeout            EnumDBErrorType = "Timeout"
)

// String returns the string representation of the EnumDBErrorType.
func (e EnumDBErrorType) String() string {
	return string(e)
}

// validDBErrs is a set of all valid EnumDBErrorType values.
var validDBErrs = map[EnumDBErrorType]struct{}{
	ErrDBAccessDenied:       {},
	ErrDBConnectionFailed:   {},
	ErrDBConstraintViolated: {},
	ErrDBDataOutOfRange:     {},
	ErrDBDuplicateEntry:     {},
	ErrDBForeignKeyViolated: {},
	ErrDBInvalidTransaction: {},
	ErrDBQueryInterrupted:   {},
	ErrDBRecordNotFound:     {},
	ErrDBSyntaxError:        {},
	ErrDBTimeout:            {},
}

// Valid checks whether the EnumDBErrorType is one of the defined constants.
func (e EnumDBErrorType) Valid() bool {
	_, ok := validDBErrs[e]
	return ok
}

var dbErrToHTTPStatusMap = map[EnumDBErrorType]int{
	ErrDBAccessDenied:       http.StatusForbidden,
	ErrDBConnectionFailed:   http.StatusServiceUnavailable,
	ErrDBConstraintViolated: http.StatusBadRequest,
	ErrDBDataOutOfRange:     http.StatusBadRequest,
	ErrDBDuplicateEntry:     http.StatusConflict,
	ErrDBForeignKeyViolated: http.StatusConflict,
	ErrDBInvalidTransaction: http.StatusServiceUnavailable,
	ErrDBQueryInterrupted:   http.StatusServiceUnavailable,
	ErrDBRecordNotFound:     http.StatusNotFound,
	ErrDBSyntaxError:        http.StatusBadRequest,
	ErrDBTimeout:            http.StatusGatewayTimeout,
}

// HTTPStatus DBErrToHTTPStatus translates the EnumDBErrorType to an HTTP status code.
func (e EnumDBErrorType) HTTPStatus() int {
	if status, ok := dbErrToHTTPStatusMap[e]; ok {
		return status
	}
	return http.StatusInternalServerError
}
