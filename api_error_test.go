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
	"strings"
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
	expectedAPIError := APIError{
		CallerID:    "CallerID",
		CallerType:  "CallerTYpe",
		InnerError:  errors.New("InnerError"),
		Message:     "Message",
		Method:      http.MethodGet,
		MultiParams: map[string][]string{"multiKey": {"multiVal1", "multiVal2"}},
		OwnerID:     "OwnerID",
		OwnerType:   "OwnerType",
		Path:        "Path",
		PathParams:  map[string]string{"pathKey1": "pathVal1", "pathKey2": "pathVal2"},
		QueryParams: map[string]string{"queryKey1": "queryVal1", "queryKey2": "queryVal2"},
		RequestID:   "RequestID",
		StatusCode:  500,
		execContext: func() execContext { return getExecContext() }(),
	}

	// Define the expected string output from the Error() method
	expectedString := "[APIError] 500 - Message at Path + GET: InnerError"

	// Get the actual error string from the APIError instance
	errString := expectedAPIError.Error()

	// Assert that the expected string matches the actual error string
	assert.Equal(t, expectedString, errString)
}

func TestLogNewAPIErr(t *testing.T) {
	setupAPIErrorFileLogger()
	defer tearDownAPIFileLogger()

	rawAPIError := GenerateRandomAPIError()

	loggedAPIError := LogAPIErr(rawAPIError)

	// Make sure to sync and close the log file to ensure all log entries are written.
	require.NoError(t, APILogFile.Sync())
	require.NoError(t, APILogFile.Close())

	// Use errors.As to unwrap the error and verify that loggedAPIError is of type *APIError
	var unwrappedAPIErr *APIError
	require.True(t, errors.As(loggedAPIError, &unwrappedAPIErr), "Error is not of type *APIError")

	t.Run("verify unwrappedAPIErr has all of its fields set correctly", func(t *testing.T) {
		assert.NotEqual(t, rawAPIError.Line, unwrappedAPIErr.Line) // these are called on different line numbers so should be different
		assert.Equal(t, DefaultAPIErrorStatusCode, unwrappedAPIErr.StatusCode)

		assert.Equal(t, rawAPIError.MultiParams, unwrappedAPIErr.MultiParams)
		assert.Equal(t, rawAPIError.PathParams, unwrappedAPIErr.PathParams)
		assert.Equal(t, rawAPIError.QueryParams, unwrappedAPIErr.QueryParams)

		assert.Equal(t, rawAPIError.CallerID, unwrappedAPIErr.CallerID)
		assert.Equal(t, rawAPIError.CallerType, unwrappedAPIErr.CallerType)
		assert.True(t, strings.HasSuffix(unwrappedAPIErr.File, "zerolog-json-structed-logs/api_error_test.go"))
		assert.True(t, strings.HasSuffix(unwrappedAPIErr.Function, "zerolog-json-structured-logs.TestLogNewAPIErr"))
		assert.Equal(t, DefaultAPIErrorMessage, unwrappedAPIErr.Message)
		assert.Equal(t, rawAPIError.Method, unwrappedAPIErr.Method)
		assert.Equal(t, rawAPIError.OwnerID, unwrappedAPIErr.OwnerID)
		assert.Equal(t, rawAPIError.OwnerType, unwrappedAPIErr.OwnerType)
		assert.Equal(t, rawAPIError.Path, unwrappedAPIErr.Path)
		assert.Equal(t, rawAPIError.RequestID, unwrappedAPIErr.RequestID)

		assert.Equal(t, rawAPIError.InnerError, unwrappedAPIErr.InnerError)
		assert.EqualError(t, rawAPIError.InnerError, unwrappedAPIErr.InnerError.Error())
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
			assert.Equal(t, float64(unwrappedAPIErr.Line), zeroLogJSONItem.ErrorAsJSON[LineKey]) // you get a float64 when unmarshalling a number into interface{} for safety
			assert.Equal(t, float64(unwrappedAPIErr.StatusCode), zeroLogJSONItem.ErrorAsJSON[StatusCodeKey])

			assert.Equal(t, unwrappedAPIErr.MultiParams, uneraseMapStringArray(zeroLogJSONItem.ErrorAsJSON[MultiParamsKey].(map[string]any)))
			assert.Equal(t, unwrappedAPIErr.PathParams, uneraseMapString(zeroLogJSONItem.ErrorAsJSON[PathParamsKey].(map[string]any)))
			assert.Equal(t, unwrappedAPIErr.QueryParams, uneraseMapString(zeroLogJSONItem.ErrorAsJSON[QueryParamsKey].(map[string]any)))

			assert.Equal(t, unwrappedAPIErr.CallerID, zeroLogJSONItem.ErrorAsJSON[CallerIDKey])
			assert.Equal(t, unwrappedAPIErr.CallerType, zeroLogJSONItem.ErrorAsJSON[CallerTypeKey])
			assert.Equal(t, unwrappedAPIErr.File, zeroLogJSONItem.ErrorAsJSON[FileKey])
			assert.Equal(t, unwrappedAPIErr.Function, zeroLogJSONItem.ErrorAsJSON[FunctionKey])
			assert.Equal(t, unwrappedAPIErr.Message, zeroLogJSONItem.ErrorAsJSON[MessageKey])
			assert.Equal(t, unwrappedAPIErr.Method, zeroLogJSONItem.ErrorAsJSON[MethodKey])
			assert.Equal(t, unwrappedAPIErr.OwnerID, zeroLogJSONItem.ErrorAsJSON[OwnerIDKey])
			assert.Equal(t, unwrappedAPIErr.OwnerType, zeroLogJSONItem.ErrorAsJSON[OwnerTypeKey])
			assert.Equal(t, unwrappedAPIErr.Path, zeroLogJSONItem.ErrorAsJSON[PathKey])
			assert.Equal(t, unwrappedAPIErr.RequestID, zeroLogJSONItem.ErrorAsJSON[RequestIDKey])

			assert.Equal(t, http.StatusText(unwrappedAPIErr.StatusCode), zeroLogJSONItem.ErrorAsJSON[StatusTextKey])
			assert.Equal(t, unwrappedAPIErr.InnerError.Error(), zeroLogJSONItem.ErrorAsJSON[InternalErrorKey]) // this is the original, top level error that DatabaseError wrapped such as a SQLError

			// check for the zerolog standard values - this is critical for testing formats and outputs for things like time and level
			assert.Equal(t, zerolog.ErrorLevel.String(), zeroLogJSONItem.Level)
			assert.Equal(t, DefaultAPIErrorMessage, zeroLogJSONItem.Message)
			assert.Equal(t, staticNowFunc(), zeroLogJSONItem.Time)
		})

		t.Run("verify that ErrorAsJSON is well formed", func(t *testing.T) {
			apiErrEntryLogValues, ok := jsonLogContents[ZLObjectKey].(map[string]any)
			require.True(t, ok, fmt.Sprintf("%s field should be a JSON object.", ZLObjectKey))

			t.Run("verify that apiErrEntryLogValues has all of its properties and values set correctly", func(t *testing.T) {
				assert.Equal(t, float64(unwrappedAPIErr.Line), apiErrEntryLogValues[LineKey]) // you get a float64 when unmarshalling a number into interface{} for safety
				assert.Equal(t, float64(unwrappedAPIErr.StatusCode), apiErrEntryLogValues[StatusCodeKey])

				assert.Equal(t, unwrappedAPIErr.MultiParams, uneraseMapStringArray(apiErrEntryLogValues[MultiParamsKey].(map[string]any)))
				assert.Equal(t, unwrappedAPIErr.PathParams, uneraseMapString(apiErrEntryLogValues[PathParamsKey].(map[string]any)))
				assert.Equal(t, unwrappedAPIErr.QueryParams, uneraseMapString(apiErrEntryLogValues[QueryParamsKey].(map[string]any)))

				assert.Equal(t, unwrappedAPIErr.CallerID, apiErrEntryLogValues[CallerIDKey])
				assert.Equal(t, unwrappedAPIErr.CallerType, apiErrEntryLogValues[CallerTypeKey])
				assert.Equal(t, unwrappedAPIErr.File, apiErrEntryLogValues[FileKey])
				assert.Equal(t, unwrappedAPIErr.Function, apiErrEntryLogValues[FunctionKey])
				assert.Equal(t, unwrappedAPIErr.Message, apiErrEntryLogValues[MessageKey])
				assert.Equal(t, unwrappedAPIErr.Method, apiErrEntryLogValues[MethodKey])
				assert.Equal(t, unwrappedAPIErr.OwnerID, apiErrEntryLogValues[OwnerIDKey])
				assert.Equal(t, unwrappedAPIErr.OwnerType, apiErrEntryLogValues[OwnerTypeKey])
				assert.Equal(t, unwrappedAPIErr.Path, apiErrEntryLogValues[PathKey])
				assert.Equal(t, unwrappedAPIErr.RequestID, apiErrEntryLogValues[RequestIDKey])

				assert.Equal(t, http.StatusText(unwrappedAPIErr.StatusCode), apiErrEntryLogValues[StatusTextKey])
				assert.Equal(t, unwrappedAPIErr.InnerError.Error(), apiErrEntryLogValues[InternalErrorKey]) // this is the original, top level error that DatabaseError wrapped such as a SQLError
			})

			t.Run("verify that struct embedding is working correctly", func(t *testing.T) {
				assert.Nil(t, apiErrEntryLogValues["exec_context"]) // struct embedding means this will NOT be in the JSON
			})
		})
	})
}
