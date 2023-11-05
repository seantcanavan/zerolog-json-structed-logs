package sl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
	"time"
)

func setupAPIErrorFileLogger() {
	// have to declare this here to prevent shadowing the outer APILogFile with :=
	var err error

	if _, err = os.Stat(TempFileNameAPILogs); err == nil {
		err = os.Remove(TempFileNameAPILogs)
		if err != nil {
			panic(fmt.Sprintf("Could not remove existing temp file: %s", err))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		// File does not exist, which is not an error in this case,
		// but any other error accessing the file system should be reported.
		panic(fmt.Sprintf("Error checking for temp file existence: %s", err))
	}

	APILogFile, err = os.CreateTemp("", TempFileNameAPILogs)
	if err != nil {
		panic(fmt.Sprintf("err is not nil: %s", err))
	}

	// Configure zerolog to use RFC3339Nano time for its output
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Configure zerolog to use a static now function for timestamp calculations so we can verify the timestamp later
	zerolog.TimestampFunc = staticNowFunc

	// Configure zerolog to write to the temp file so we can easily capture the output
	log.Logger = zerolog.New(APILogFile).With().Timestamp().Logger()
	zerolog.DisableSampling(true)
}

func tearDownAPIFileLogger() {
	err := os.Remove(APILogFile.Name())
	if err != nil {
		panic(fmt.Sprintf("err is not nil: %s", err))
	}
}

func TestAPIError_Error(t *testing.T) {
	// Set up fake values for the expected API error
	expectedAPIError := apiError{
		APIEndpoint:   "/test/endpoint",
		CallerID:      "caller123",
		InternalError: errors.New("internal server error"),
		Message:       "An error occurred",
		RequestID:     "req-123",
		StatusCode:    500,
		UserID:        "user123",
		execContext:   func() execContext { return getExecContext() }(),
	}

	// Define the expected string output from the Error() method
	expectedString := "[apiError] 500 - An error occurred at /test/endpoint: internal server error"

	// Get the actual error string from the apiError instance
	errString := expectedAPIError.Error()

	// Assert that the expected string matches the actual error string
	assert.Equal(t, expectedString, errString)
}

func TestLogNewAPIErr(t *testing.T) {
	setupAPIErrorFileLogger()
	defer tearDownAPIFileLogger()

	// this gets propagated up to the LogItem
	message := "time.Parse failed"

	expectedAPIErr := apiError{
		APIEndpoint:   "/test/endpoint",
		CallerID:      "caller-123",
		InternalError: errors.New("internal server error"),
		Message:       message,
		RequestID:     "req-123",
		StatusCode:    http.StatusTeapot,
		UserID:        "user-123",

		// we have to wrap this so the call stack is at the same depth as LogNewAPIErr below
		execContext: func() execContext { return getExecContext() }(),
	}

	newAPIErr := LogNewAPIErr(NewAPIErr{
		APIEndpoint:   expectedAPIErr.APIEndpoint,
		CallerID:      expectedAPIErr.CallerID,
		InternalError: expectedAPIErr.InternalError,
		Message:       expectedAPIErr.Message,
		RequestID:     expectedAPIErr.RequestID,
		StatusCode:    expectedAPIErr.StatusCode,
		UserID:        expectedAPIErr.UserID,
	})

	// Make sure to sync and close the log file to ensure all log entries are written.
	require.NoError(t, APILogFile.Sync())
	require.NoError(t, APILogFile.Close())

	// Use errors.As to unwrap the error and verify that newAPIErr is of type *apiError
	var unwrappedAPIErr *apiError
	require.True(t, errors.As(newAPIErr, &unwrappedAPIErr), "Error is not of type *apiError")

	t.Run("verify unwrappedAPIErr has all of its fields set correctly", func(t *testing.T) {
		assert.Equal(t, expectedAPIErr.APIEndpoint, unwrappedAPIErr.APIEndpoint)
		assert.Equal(t, expectedAPIErr.CallerID, unwrappedAPIErr.CallerID)
		assert.Equal(t, expectedAPIErr.File, unwrappedAPIErr.File)
		assert.Equal(t, expectedAPIErr.Function, unwrappedAPIErr.Function)
		assert.Equal(t, expectedAPIErr.InternalError, unwrappedAPIErr.InternalError)
		assert.NotEqual(t, expectedAPIErr.Line, unwrappedAPIErr.Line) // these are called on different line numbers so should be different
		assert.Equal(t, expectedAPIErr.Message, unwrappedAPIErr.Message)
		assert.Equal(t, expectedAPIErr.RequestID, unwrappedAPIErr.RequestID)
		assert.Equal(t, expectedAPIErr.StatusCode, unwrappedAPIErr.StatusCode)
		assert.Equal(t, expectedAPIErr.UserID, unwrappedAPIErr.UserID)
		assert.EqualError(t, expectedAPIErr.InternalError, unwrappedAPIErr.InternalError.Error())
	})

	t.Run("verify that jsonLogContents is well formed", func(t *testing.T) {
		// Read the log file's logFileJSONContents for assertion.
		logFileJSONContents, err := os.ReadFile(APILogFile.Name())
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
			assert.Equal(t, unwrappedAPIErr.APIEndpoint, zeroLogJSONItem.ErrorAsJSON["apiEndpoint"])
			assert.Equal(t, unwrappedAPIErr.CallerID, zeroLogJSONItem.ErrorAsJSON["callerId"])
			assert.Equal(t, unwrappedAPIErr.File, zeroLogJSONItem.ErrorAsJSON["file"])
			assert.Equal(t, unwrappedAPIErr.Function, zeroLogJSONItem.ErrorAsJSON["function"])
			assert.Equal(t, unwrappedAPIErr.InternalError.Error(), zeroLogJSONItem.ErrorAsJSON["internalError"]) // this is the original, top level error that databaseError wrapped such as a SQLError
			assert.Equal(t, float64(unwrappedAPIErr.Line), zeroLogJSONItem.ErrorAsJSON["line"])                  // you get a float64 when unmarshalling a number into interface{} for safety
			assert.Equal(t, unwrappedAPIErr.Message, zeroLogJSONItem.ErrorAsJSON["message"])
			assert.Equal(t, unwrappedAPIErr.RequestID, zeroLogJSONItem.ErrorAsJSON["requestId"])
			assert.Equal(t, float64(unwrappedAPIErr.StatusCode), zeroLogJSONItem.ErrorAsJSON["statusCode"])
			assert.Equal(t, unwrappedAPIErr.UserID, zeroLogJSONItem.ErrorAsJSON["userId"])

			// check for the zerolog standard values - this is critical for testing formats and outputs for things like time and level
			assert.Equal(t, zerolog.ErrorLevel.String(), zeroLogJSONItem.Level)
			assert.Equal(t, message, zeroLogJSONItem.Message)
			assert.Equal(t, staticNowFunc(), zeroLogJSONItem.Time)
		})

		t.Run("verify that ErrorAsJSON is well formed", func(t *testing.T) {
			apiErrEntryLogValues, ok := jsonLogContents[ZLObjectKey].(map[string]any)
			require.True(t, ok, fmt.Sprintf("%s field should be a JSON object.", ZLObjectKey))

			t.Run("verify that apiErrEntryLogValues has all of its properties and values set correctly", func(t *testing.T) {
				assert.Equal(t, unwrappedAPIErr.APIEndpoint, apiErrEntryLogValues["apiEndpoint"])
				assert.Equal(t, unwrappedAPIErr.CallerID, apiErrEntryLogValues["callerId"])
				assert.Equal(t, unwrappedAPIErr.File, apiErrEntryLogValues["file"])
				assert.Equal(t, unwrappedAPIErr.Function, apiErrEntryLogValues["function"])
				assert.Equal(t, unwrappedAPIErr.InternalError.Error(), apiErrEntryLogValues["internalError"]) // this is the original, top level error that databaseError wrapped such as a SQLError
				assert.Equal(t, float64(unwrappedAPIErr.Line), apiErrEntryLogValues["line"])                  // you get a float64 when unmarshalling a number into interface{} for safety
				assert.Equal(t, unwrappedAPIErr.Message, apiErrEntryLogValues["message"])
				assert.Equal(t, unwrappedAPIErr.RequestID, apiErrEntryLogValues["requestId"])
				assert.Equal(t, float64(unwrappedAPIErr.StatusCode), apiErrEntryLogValues["statusCode"])
				assert.Equal(t, unwrappedAPIErr.UserID, apiErrEntryLogValues["userId"])
			})

			t.Run("verify that struct embedding is working correctly", func(t *testing.T) {
				assert.Nil(t, apiErrEntryLogValues["exec_context"]) // struct embedding means this will NOT be in the JSON
			})
		})
	})
}
