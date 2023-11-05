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
		{"ErrDBSomeOtherDatabaseError", http.StatusInternalServerError},
		{ErrDBAccessDenied, http.StatusForbidden},
		{ErrDBConnectionFailed, http.StatusServiceUnavailable},
		{ErrDBConstraintViolated, http.StatusBadRequest},
		{ErrDBDataOutOfRange, http.StatusBadRequest},
		{ErrDBDuplicateEntry, http.StatusConflict},
		{ErrDBForeignKeyViolated, http.StatusConflict},
		{ErrDBInvalidTransaction, http.StatusServiceUnavailable},
		{ErrDBQueryInterrupted, http.StatusServiceUnavailable},
		{ErrDBRecordNotFound, http.StatusNotFound},
		{ErrDBSyntaxError, http.StatusBadRequest},
		{ErrDBTimeout, http.StatusGatewayTimeout},
	}

	for _, tc := range testCases {
		t.Run(string(tc.dbErrEnum), func(t *testing.T) {
			status := tc.dbErrEnum.HTTPStatus()
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

// TestEnumDBError_String tests the String method of the EnumDBErrorType type.
func TestEnumDBError_String(t *testing.T) {
	testCases := []struct {
		enumDBError    EnumDBErrorType
		expectedString string
	}{
		{"ErrSomeRandomError", "ErrSomeRandomError"},
		{ErrDBAccessDenied, "Access Denied"},
		{ErrDBConnectionFailed, "Connection Failed"},
		{ErrDBConstraintViolated, "Constraint Violation"},
		{ErrDBDataOutOfRange, "Data Out of Range"},
		{ErrDBDuplicateEntry, "Duplicate Entry"},
		{ErrDBForeignKeyViolated, "Foreign Key Violation"},
		{ErrDBInvalidTransaction, "Invalid Transaction"},
		{ErrDBQueryInterrupted, "Query Interrupted"},
		{ErrDBRecordNotFound, "Record Not Found"},
		{ErrDBSyntaxError, "Syntax Error"},
		{ErrDBTimeout, "Timeout"},
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
		{ErrDBForeignKeyViolated, true},
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
