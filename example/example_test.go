package example

import (
	"errors"
	"fmt"
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

	apiErr := sl.GenerateNonRandomAPIError()
	apiErr.InnerError = fmt.Errorf("wrapping db error %w", expectedDBError)
	apiErr.StatusCode = sl.ErrDBConnectionFailed.HTTPStatus()

	// Define the expected APIError
	expectedAPIError := sl.LogAPIErr(apiErr)

	// Wrap the DatabaseError in an APIError
	wrappedAPIError := wrapDatabaseError()
	require.NotNil(t, wrappedAPIError)

	var unwrappedExpectedAPIError *sl.APIError
	require.True(t, errors.As(expectedAPIError, &unwrappedExpectedAPIError))

	var unwrappedExpectedDBError *sl.DatabaseError
	require.True(t, errors.As(unwrappedExpectedAPIError.InnerError, &unwrappedExpectedDBError))

	// Unwrap the error to assert on the API error
	var unwrappedAPIErr *sl.APIError
	require.True(t, errors.As(wrappedAPIError, &unwrappedAPIErr))

	var dbErr *sl.DatabaseError
	require.True(t, errors.As(unwrappedAPIErr.InnerError, &dbErr))

	// Assert the properties of the APIError itself
	assert.Equal(t, unwrappedExpectedAPIError.Path, unwrappedAPIErr.Path)
	assert.Equal(t, unwrappedExpectedAPIError.CallerID, unwrappedAPIErr.CallerID)
	assert.Equal(t, unwrappedExpectedAPIError.Message, unwrappedAPIErr.Message)
	assert.Equal(t, unwrappedExpectedAPIError.RequestID, unwrappedAPIErr.RequestID)
	assert.Equal(t, unwrappedExpectedAPIError.StatusCode, unwrappedAPIErr.StatusCode)
	assert.Equal(t, unwrappedExpectedAPIError.OwnerID, unwrappedAPIErr.OwnerID)

	// Unwrap the internal error of the APIError to get the DatabaseError
	var unwrappedDBErr *sl.DatabaseError
	require.True(t, errors.As(unwrappedAPIErr.InnerError, &unwrappedDBErr))

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
