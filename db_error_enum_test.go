package sl

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestEnumDBErr_HTTPStatus(t *testing.T) {
	testCases := []struct {
		dbErrEnum      EnumDBErrorType
		expectedStatus int
	}{
		{ErrDBDuplicateEntry, http.StatusConflict},
		{ErrDBRecordNotFound, http.StatusNotFound},
		{ErrDBConnectionFailed, http.StatusServiceUnavailable},
		{ErrDBQueryInterrupted, http.StatusServiceUnavailable},
		{ErrDBInvalidTransaction, http.StatusServiceUnavailable},
		{ErrDBConstraintViolated, http.StatusBadRequest},
		{ErrDBDataOutOfRange, http.StatusBadRequest},
		{ErrDBSyntaxError, http.StatusBadRequest},
		{ErrDBAccessDenied, http.StatusForbidden},
		{ErrDBTimeout, http.StatusGatewayTimeout},
		{"ErrDBSomeOtherDatabaseError", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(string(tc.dbErrEnum), func(t *testing.T) {
			status := tc.dbErrEnum.HTTPStatus()
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestEnumDBErr_HTTPStatusAndText(t *testing.T) {
	testCases := []struct {
		enumDBErr      EnumDBErrorType
		expectedStatus int
		expectedText   string
	}{
		{ErrDBAccessDenied, http.StatusForbidden, "Forbidden"},
		{ErrDBConnectionFailed, http.StatusServiceUnavailable, "Service Unavailable"},
		{ErrDBConstraintViolated, http.StatusBadRequest, "Bad Request"},
		{ErrDBDataOutOfRange, http.StatusBadRequest, "Bad Request"},
		{ErrDBDuplicateEntry, http.StatusConflict, "Conflict"},
		{ErrDBInvalidTransaction, http.StatusServiceUnavailable, "Service Unavailable"},
		{ErrDBQueryInterrupted, http.StatusServiceUnavailable, "Service Unavailable"},
		{ErrDBRecordNotFound, http.StatusNotFound, "Not Found"},
		{ErrDBSyntaxError, http.StatusBadRequest, "Bad Request"},
		{ErrDBTimeout, http.StatusGatewayTimeout, "Gateway Timeout"},
		{"undefined_error", http.StatusInternalServerError, "Internal Server Error"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.enumDBErr), func(t *testing.T) {
			status, text := tc.enumDBErr.HTTPStatusAndText()
			assert.Equal(t, tc.expectedStatus, status)
			assert.Equal(t, tc.expectedText, text)
		})
	}
}

// TestEnumDBError_String tests the String method of the EnumDBErrorType type.
func TestEnumDBError_String(t *testing.T) {
	testCases := []struct {
		enumDBError    EnumDBErrorType
		expectedString string
	}{
		{ErrDBAccessDenied, "Access Denied"},
		{ErrDBConnectionFailed, "Connection Failed"},
		{ErrDBConstraintViolated, "Constraint Violation"},
		{ErrDBDataOutOfRange, "Data Out of Range"},
		{ErrDBDuplicateEntry, "Duplicate Entry"},
		{ErrDBInvalidTransaction, "Invalid Transaction"},
		{ErrDBQueryInterrupted, "Query Interrupted"},
		{ErrDBRecordNotFound, "Record Not Found"},
		{ErrDBSyntaxError, "Syntax Error"},
		{ErrDBTimeout, "Timeout"},
		{"undefined_error", "undefined_error"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.enumDBError), func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.enumDBError.String())
		})
	}
}

func TestEnumDBError_Valid(t *testing.T) {
	testCases := []struct {
		enumDBError   EnumDBErrorType
		expectedValid bool
	}{
		{ErrDBAccessDenied, true},
		{ErrDBConnectionFailed, true},
		{ErrDBConstraintViolated, true},
		{ErrDBDataOutOfRange, true},
		{ErrDBDuplicateEntry, true},
		{ErrDBInvalidTransaction, true},
		{ErrDBQueryInterrupted, true},
		{ErrDBRecordNotFound, true},
		{ErrDBSyntaxError, true},
		{ErrDBTimeout, true},
		// Add a case for an undefined error to ensure it returns false
		{"undefined_error", false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.enumDBError), func(t *testing.T) {
			valid := tc.enumDBError.Valid()
			assert.Equal(t, tc.expectedValid, valid)
		})
	}
}
