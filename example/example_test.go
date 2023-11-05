package example

import (
	"errors"
	sl "github.com/seantcanavan/zerolog-json-structured-logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWrapDatabaseError(t *testing.T) {
	// Define the expected DatabaseError
	expectedDBError := sl.LogNewDBErr(sl.NewDBErr{
		Constraint:    "pk_users",
		DBName:        "testdb",
		InternalError: errors.New("sql: no rows in result set"),
		Message:       "connection to database failed",
		Operation:     "SELECT",
		Query:         "SELECT * FROM users",
		TableName:     "users",
		Type:          sl.ErrDBConnectionFailed,
	})

	// Define the expected APIError
	expectedAPIError := sl.LogNewAPIErr(sl.NewAPIErr{
		APIEndpoint:   "/test/endpoint",
		CallerID:      "caller-123",
		Message:       "cannot get users by address",
		RequestID:     "req-123",
		InternalError: expectedDBError,
		StatusCode:    sl.ErrDBConnectionFailed.HTTPStatus(),
		UserID:        "user-123",
	})

	// Wrap the DatabaseError in an APIError
	wrappedAPIError := wrapDatabaseError()
	require.NotNil(t, wrappedAPIError)

	var unwrappedExpectedAPIError *sl.APIError
	require.True(t, errors.As(expectedAPIError, &unwrappedExpectedAPIError))

	var unwrappedExpectedDBError *sl.DatabaseError
	require.True(t, errors.As(unwrappedExpectedAPIError.InternalError, &unwrappedExpectedDBError))

	// Unwrap the error to assert on the API error
	var apiErr *sl.APIError
	require.True(t, errors.As(wrappedAPIError, &apiErr))

	var dbErr *sl.DatabaseError
	require.True(t, errors.As(apiErr.InternalError, &dbErr))

	// Assert the properties of the APIError itself
	assert.Equal(t, unwrappedExpectedAPIError.APIEndpoint, apiErr.APIEndpoint)
	assert.Equal(t, unwrappedExpectedAPIError.CallerID, apiErr.CallerID)
	assert.Equal(t, unwrappedExpectedAPIError.Message, apiErr.Message)
	assert.Equal(t, unwrappedExpectedAPIError.RequestID, apiErr.RequestID)
	assert.Equal(t, unwrappedExpectedAPIError.StatusCode, apiErr.StatusCode)
	assert.Equal(t, unwrappedExpectedAPIError.StatusText, apiErr.StatusText)
	assert.Equal(t, unwrappedExpectedAPIError.UserID, apiErr.UserID)

	// Unwrap the internal error of the APIError to get the DatabaseError
	var unwrappedDBErr *sl.DatabaseError
	require.True(t, errors.As(apiErr.InternalError, &unwrappedDBErr))

	// Assert the properties of the unwrapped DatabaseError
	assert.Equal(t, unwrappedExpectedDBError.Constraint, dbErr.Constraint)
	assert.Equal(t, unwrappedExpectedDBError.DBName, dbErr.DBName)
	assert.Equal(t, unwrappedExpectedDBError.Message, dbErr.Message)
	assert.Equal(t, unwrappedExpectedDBError.Operation, dbErr.Operation)
	assert.Equal(t, unwrappedExpectedDBError.Query, dbErr.Query)
	assert.Equal(t, unwrappedExpectedDBError.TableName, dbErr.TableName)
	assert.Equal(t, unwrappedExpectedDBError.Type, dbErr.Type)
}

// FindLemonadeStandError searches the error chain for a LemonadeStandError.
func findLemonadeStandError(err error) (*lemonadeStandError, bool) {
	var lse *lemonadeStandError
	for {
		if errors.As(err, &lse) {
			return lse, true
		}
		// If the error does not have a cause (i.e., it is not wrapped), we exit the loop.
		if err = errors.Unwrap(err); err == nil {
			break
		}
	}
	return nil, false
}
