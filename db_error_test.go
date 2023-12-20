package sl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"testing"
	"time"
)

const TempFileNameDBLogs = "testdblogs"
const TempFileNameAPILogs = "testapilogs"

var APILogFile *os.File // zerolog writes to this file so we can capture the output
var DBLogFile *os.File  // zerolog writes to this file so we can capture the output

// generate a new, random Now every execution. This helps test more permutations of dates and edge cases
var staticNow = func() time.Time {
	nowRand := rand.New(rand.NewSource(time.Now().Unix()))
	// Generating a random year, month, day, etc.
	year := nowRand.Intn(2023-2000) + 2000 // random year between 2000 and 2023
	month := time.Month(nowRand.Intn(12) + 1)
	day := nowRand.Intn(28) + 1 // to avoid issues with February, keep it up to 28
	hour := nowRand.Intn(24)
	minute := nowRand.Intn(60)
	second := nowRand.Intn(60)

	// Constructing the random date using the time.Date function
	randomDate := time.Date(year, month, day, hour, minute, second, 0, time.UTC)

	return randomDate
}()

func staticNowFunc() time.Time {
	return staticNow
}

func setupDBErrorFileLogger() {
	// have to declare this here to prevent shadowing the outer DBLogFile with :=
	var err error

	if _, err = os.Stat(TempFileNameDBLogs); err == nil {
		err = os.Remove(TempFileNameDBLogs)
		if err != nil {
			panic(fmt.Sprintf("Could not remove existing temp file: %s", err))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		// File does not exist, which is not an error in this case,
		// but any other error accessing the file system should be reported.
		panic(fmt.Sprintf("Error checking for temp file existence: %s", err))
	}

	DBLogFile, err = os.CreateTemp("", TempFileNameDBLogs)
	if err != nil {
		panic(fmt.Sprintf("err is not nil: %s", err))
	}

	// Configure zerolog to use RFC3339Nano time for its output
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Configure zerolog to use a static now function for timestamp calculations so we can verify the timestamp later
	zerolog.TimestampFunc = staticNowFunc

	// Configure zerolog to write to the temp file so we can easily capture the output
	log.Logger = zerolog.New(DBLogFile).With().Timestamp().Logger()
	zerolog.DisableSampling(true)
}

func tearDownDatabaseFileLogger() {
	err := os.Remove(DBLogFile.Name())
	if err != nil {
		panic(fmt.Sprintf("err is not nil: %s", err))
	}
}

func TestDBError_Error(t *testing.T) {
	expectedDBErr := DatabaseError{
		// Assume these values are what you expect to see after the operation.
		Constraint:    "Constraint",
		DBName:        "DBName",
		InternalError: errors.New("InnerError"),
		Message:       "Message",
		Operation:     "Operation",
		Query:         "Query",
		TableName:     "TableName",
		Type:          ErrDBConnectionFailed,

		// we have to wrap this so the call stack is at the same depth as LogNewDBErr below
		execContext: func() execContext { return getExecContext() }(),
	}

	errString := expectedDBErr.Error()

	expectedString := "[DatabaseError] Operation operation on DBName.TableName with query: Query - Message - InnerError"

	assert.Equal(t, expectedString, errString)
}

func TestLogNewDBErr(t *testing.T) {
	setupDBErrorFileLogger()
	defer tearDownDatabaseFileLogger()

	// this gets propagated up to the LogItem
	message := "no users found"

	expectedDBErr := DatabaseError{
		// Assume these values are what you expect to see after the operation.
		Constraint:    "pk_users",
		DBName:        "testdb",
		InternalError: fmt.Errorf("wrapping error %w", errors.New("sql: no rows in result set")),
		Message:       message,
		Operation:     "SELECT",
		Query:         "SELECT * FROM users",
		TableName:     "users",
		Type:          ErrDBConnectionFailed,

		// we have to wrap this so the call stack is at the same depth as LogNewDBErr below
		execContext: func() execContext { return getExecContext() }(),
	}

	newDBErr := LogNewDBErr(NewDBErr{ // Call LogNewDBErr to log the error to the temp file
		Constraint:    expectedDBErr.Constraint,
		DBName:        expectedDBErr.DBName,
		InternalError: errors.New("sql: no rows in result set"),
		Message:       expectedDBErr.Message,
		Operation:     expectedDBErr.Operation,
		Query:         expectedDBErr.Query,
		TableName:     expectedDBErr.TableName,
		Type:          expectedDBErr.Type,
	})

	// Make sure to sync and close the log file to ensure all log entries are written.
	require.NoError(t, DBLogFile.Sync())
	require.NoError(t, DBLogFile.Close())

	// Use errors.As to unwrap the error and verify that newDBErr is of type *DatabaseError
	var unwrappedNewDBErr *DatabaseError
	require.True(t, errors.As(newDBErr, &unwrappedNewDBErr), "Error is not of type *DatabaseError")

	t.Run("verify unwrappedNewDBErr has all of its fields set correctly", func(t *testing.T) {
		assert.Equal(t, expectedDBErr.Constraint, unwrappedNewDBErr.Constraint)
		assert.Equal(t, expectedDBErr.DBName, unwrappedNewDBErr.DBName)
		assert.Equal(t, expectedDBErr.File, unwrappedNewDBErr.File)
		assert.Equal(t, expectedDBErr.Function, unwrappedNewDBErr.Function)
		assert.Equal(t, expectedDBErr.InternalError, unwrappedNewDBErr.InternalError)
		assert.NotEqual(t, expectedDBErr.Line, unwrappedNewDBErr.Line) // these are called on different line numbers so should be different
		assert.Equal(t, expectedDBErr.Message, unwrappedNewDBErr.Message)
		assert.Equal(t, expectedDBErr.Operation, unwrappedNewDBErr.Operation)
		assert.Equal(t, expectedDBErr.Query, unwrappedNewDBErr.Query)
		assert.Equal(t, expectedDBErr.TableName, unwrappedNewDBErr.TableName)
		assert.Equal(t, expectedDBErr.Type, unwrappedNewDBErr.Type)
		assert.EqualError(t, expectedDBErr.InternalError, unwrappedNewDBErr.InternalError.Error())
	})

	t.Run("verify that jsonLogContents is well formed", func(t *testing.T) {
		// Read the log file's logFileJSONContents for assertion.
		logFileJSONContents, err := os.ReadFile(DBLogFile.Name())
		require.NoError(t, err)

		// Unmarshal logFileJSONContents into a generic map[string]any
		var jsonLogContents map[string]any
		require.NoError(t, json.Unmarshal(logFileJSONContents, &jsonLogContents), "Error unmarshalling log logFileJSONContents")
		require.NotEmpty(t, jsonLogContents, "Log file should contain at least one entry.")
		require.NotNil(t, jsonLogContents[ZLObjectKey], fmt.Sprintf("Log entry should contain '%s' field.", ZLObjectKey))

		t.Run("verify that jsonLogContents unmarshals into an instance of ZLJSONItem", func(t *testing.T) {
			var zeroLogJSONItem ZLJSONItem
			require.NoError(t, json.Unmarshal(logFileJSONContents, &zeroLogJSONItem), "json.Unmarshal should not have produced an error")

			// check for the error values embedded in the top-level logging struct
			assert.Equal(t, unwrappedNewDBErr.Constraint, zeroLogJSONItem.ErrorAsJSON["constraint"])
			assert.Equal(t, unwrappedNewDBErr.DBName, zeroLogJSONItem.ErrorAsJSON["dbName"])
			assert.Equal(t, unwrappedNewDBErr.File, zeroLogJSONItem.ErrorAsJSON["file"])
			assert.Equal(t, unwrappedNewDBErr.Function, zeroLogJSONItem.ErrorAsJSON["function"])
			assert.Equal(t, unwrappedNewDBErr.InternalError.Error(), zeroLogJSONItem.ErrorAsJSON["internalError"]) // this is the original, top level error that DatabaseError wrapped such as a SQLError
			assert.Equal(t, float64(unwrappedNewDBErr.Line), zeroLogJSONItem.ErrorAsJSON["line"])                  // you get a float64 when unmarshalling a number into interface{} for safety
			assert.Equal(t, unwrappedNewDBErr.Message, zeroLogJSONItem.ErrorAsJSON["message"])
			assert.Equal(t, unwrappedNewDBErr.Operation, zeroLogJSONItem.ErrorAsJSON["operation"])
			assert.Equal(t, unwrappedNewDBErr.Query, zeroLogJSONItem.ErrorAsJSON["query"])
			assert.Equal(t, unwrappedNewDBErr.TableName, zeroLogJSONItem.ErrorAsJSON["tableName"])

			// check for the zerolog standard values - this is critical for testing formats and outputs for things like time and level
			assert.Equal(t, zerolog.ErrorLevel.String(), zeroLogJSONItem.Level)
			assert.Equal(t, message, zeroLogJSONItem.Message)
			assert.Equal(t, staticNowFunc(), zeroLogJSONItem.Time)
		})

		t.Run("verify that ErrorAsJSON is well formed", func(t *testing.T) {
			dbErrEntryLogValues, ok := jsonLogContents[ZLObjectKey].(map[string]any)
			require.True(t, ok, fmt.Sprintf("%s field should be a JSON object.", ZLObjectKey))

			t.Run("verify that dbErrEntryLogValues has all of its properties and values set correctly", func(t *testing.T) {
				assert.Equal(t, unwrappedNewDBErr.Constraint, dbErrEntryLogValues["constraint"])
				assert.Equal(t, unwrappedNewDBErr.DBName, dbErrEntryLogValues["dbName"])
				assert.Equal(t, unwrappedNewDBErr.File, dbErrEntryLogValues["file"])
				assert.Equal(t, unwrappedNewDBErr.Function, dbErrEntryLogValues["function"])
				assert.Equal(t, unwrappedNewDBErr.InternalError.Error(), dbErrEntryLogValues["internalError"]) // this is the original, top level error that DatabaseError wrapped such as a SQLError
				assert.Equal(t, float64(unwrappedNewDBErr.Line), dbErrEntryLogValues["line"])                  // you get a float64 when unmarshalling a number into interface{} for safety
				assert.Equal(t, unwrappedNewDBErr.Message, dbErrEntryLogValues["message"])
				assert.Equal(t, unwrappedNewDBErr.Operation, dbErrEntryLogValues["operation"])
				assert.Equal(t, unwrappedNewDBErr.Query, dbErrEntryLogValues["query"])
				assert.Equal(t, unwrappedNewDBErr.TableName, dbErrEntryLogValues["tableName"])
			})

			t.Run("verify that struct embedding is working correctly", func(t *testing.T) {
				assert.Nil(t, dbErrEntryLogValues["exec_context"]) // struct embedding means this will NOT be in the JSON
			})
		})
	})
}

func TestFindLastDatabaseError(t *testing.T) {
	firstError := LogNewDBErr(NewDBErr{
		Constraint:    "pk_users_id",
		DBName:        "usersdb",
		InternalError: fmt.Errorf("primary key violation"),
		Message:       "duplicate entry for primary key",
		Operation:     "INSERT",
		Query:         "INSERT INTO users (id, name) VALUES (1, 'John Doe')",
		TableName:     "users",
		Type:          ErrDBConstraintViolated,
	})

	secondError := LogNewDBErr(NewDBErr{
		Constraint:    "fk_orders_user_id",
		DBName:        "ordersdb",
		InternalError: firstError,
		Message:       "invalid foreign key",
		Operation:     "UPDATE",
		Query:         "UPDATE orders SET user_id = 2 WHERE order_id = 99",
		TableName:     "orders",
		Type:          ErrDBForeignKeyViolated,
	})

	// Test
	outermostErr := FindOutermostDatabaseError(secondError)
	require.NotNil(t, outermostErr)

	// Unwrap the error to assert on the last database error in the chain
	var outermostDBErr *DatabaseError
	require.True(t, errors.As(outermostErr, &outermostDBErr))

	var secondErrorUnwrapped *DatabaseError
	require.True(t, errors.As(outermostErr, &secondErrorUnwrapped))

	var firstErrorUnwrapped *DatabaseError
	require.True(t, errors.As(secondErrorUnwrapped.InternalError, &firstErrorUnwrapped))

	// Compare the outermost error returned to the second error defined
	assert.Equal(t, secondErrorUnwrapped.Constraint, outermostDBErr.Constraint)
	assert.Equal(t, secondErrorUnwrapped.DBName, outermostDBErr.DBName)
	assert.Equal(t, secondErrorUnwrapped.File, outermostDBErr.File)
	assert.Equal(t, secondErrorUnwrapped.Function, outermostDBErr.Function)
	assert.Equal(t, secondErrorUnwrapped.Line, outermostDBErr.Line)
	assert.Equal(t, secondErrorUnwrapped.Message, outermostDBErr.Message)
	assert.Equal(t, secondErrorUnwrapped.Operation, outermostDBErr.Operation)
	assert.Equal(t, secondErrorUnwrapped.Query, outermostDBErr.Query)
	assert.Equal(t, secondErrorUnwrapped.TableName, outermostDBErr.TableName)
	assert.Equal(t, secondErrorUnwrapped.Type, outermostDBErr.Type)

	// Compare the error wrapped by the outermost error to the first error defined
	assert.Equal(t, firstErrorUnwrapped.Constraint, firstErrorUnwrapped.Constraint)
	assert.Equal(t, firstErrorUnwrapped.DBName, firstErrorUnwrapped.DBName)
	assert.Equal(t, firstErrorUnwrapped.File, firstErrorUnwrapped.File)
	assert.Equal(t, firstErrorUnwrapped.Function, firstErrorUnwrapped.Function)
	assert.Equal(t, firstErrorUnwrapped.Line, firstErrorUnwrapped.Line)
	assert.Equal(t, firstErrorUnwrapped.Message, firstErrorUnwrapped.Message)
	assert.Equal(t, firstErrorUnwrapped.Operation, firstErrorUnwrapped.Operation)
	assert.Equal(t, firstErrorUnwrapped.Query, firstErrorUnwrapped.Query)
	assert.Equal(t, firstErrorUnwrapped.TableName, firstErrorUnwrapped.TableName)
	assert.Equal(t, firstErrorUnwrapped.Type, firstErrorUnwrapped.Type)
}
