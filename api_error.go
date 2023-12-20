package sl

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

const CallerIDKey = "callerId"
const CallerTypeKey = "callerType"
const FileKey = "file"
const FunctionKey = "function"
const InternalErrorKey = "internalError"
const LineKey = "line"
const MessageKey = "message"
const MethodKey = "method"
const MultiParamsKey = "multiParams"
const OwnerIDKey = "ownerId"
const OwnerTypeKey = "ownerType"
const PathKey = "path"
const PathParamsKey = "pathParams"
const QueryParamsKey = "queryParams"
const RequestIDKey = "requestId"
const StatusCodeKey = "statusCode"
const StatusTextKey = "statusText"

const DefaultAPIErrorMessage = "an API Error occurred"
const DefaultAPIErrorStatusCode = http.StatusInternalServerError

// APIError represents an error that occurred in the API layer of the application.
// It includes details like the HTTP status code and additional context.
type APIError struct {
	CallerID    string              `json:"callerId,omitempty"`
	CallerType  string              `json:"callerType,omitempty"`
	InnerError  error               `json:"innerError,omitempty"` // An internal error if it exists such as twilio.SendSMS or other integrations
	Message     string              `json:"message,omitempty"`
	Method      string              `json:"method,omitempty"`
	MultiParams map[string][]string `json:"multiParams,omitempty"`
	OwnerID     string              `json:"ownerId,omitempty"`
	OwnerType   string              `json:"ownerType,omitempty"`
	Path        string              `json:"path,omitempty"`
	PathParams  map[string]string   `json:"pathParams,omitempty"`
	QueryParams map[string]string   `json:"queryParams,omitempty"`
	RequestID   string              `json:"requestId,omitempty"`
	StatusCode  int                 `json:"statusCode,omitempty"`

	execContext `json:"execContext"` // Embedded struct
}

// Error returns the string representation of the APIError.
func (e *APIError) Error() string {
	return fmt.Sprintf("[APIError] %d - %s at %s + %s: %s", e.StatusCode, e.Message, e.Path, e.Method, e.InnerError)
}

// Unwrap provides the underlying error for use with errors.Is and errors.As functions.
func (e *APIError) Unwrap() error {
	return e.InnerError
}

func addDefaults(apiErr *APIError) {
	if apiErr.Message == "" {
		apiErr.Message = DefaultAPIErrorMessage
	}

	if apiErr.StatusCode == 0 {
		apiErr.StatusCode = DefaultAPIErrorStatusCode
	}

	if apiErr.InnerError == nil {
		apiErr.InnerError = errors.New(apiErr.Message)
	}
}

func LogAPIErrCtx(ctx context.Context, err error, message string, statusCode int) error {
	apiErr := APIError{
		CallerID:    fromCtxSafe[string](ctx, CallerIDKey),
		CallerType:  fromCtxSafe[string](ctx, CallerTypeKey),
		InnerError:  err,
		Message:     message,
		Method:      fromCtxSafe[string](ctx, MethodKey),
		MultiParams: fromCtxSafe[map[string][]string](ctx, MultiParamsKey),
		OwnerID:     fromCtxSafe[string](ctx, OwnerIDKey),
		OwnerType:   fromCtxSafe[string](ctx, OwnerTypeKey),
		Path:        fromCtxSafe[string](ctx, PathKey),
		PathParams:  fromCtxSafe[map[string]string](ctx, PathParamsKey),
		QueryParams: fromCtxSafe[map[string]string](ctx, QueryParamsKey),
		RequestID:   fromCtxSafe[string](ctx, RequestIDKey),
		StatusCode:  statusCode,
		execContext: execContext{},
	}

	addDefaults(&apiErr)

	log.Error().
		Object(ZLObjectKey, &apiErr).
		Msg(apiErr.Message)

	return &apiErr
}

func LogAPIErr(apiErr APIError) error {
	addDefaults(&apiErr)

	apiErr.execContext = getExecContext()

	log.Error().
		Object(ZLObjectKey, &apiErr).
		Msg(apiErr.Message)

	return &apiErr
}

// MarshalZerologObject allows APIError to be logged by zerolog.
func (e *APIError) MarshalZerologObject(zle *zerolog.Event) {
	zle.
		Int(LineKey, e.Line).
		Int(StatusCodeKey, e.StatusCode).
		Interface(MultiParamsKey, e.MultiParams).
		Interface(PathParamsKey, e.PathParams).
		Interface(QueryParamsKey, e.QueryParams).
		Str(CallerIDKey, e.CallerID).
		Str(CallerTypeKey, e.CallerType).
		Str(FileKey, e.File).
		Str(FunctionKey, e.Function).
		Str(MessageKey, e.Message).
		Str(MethodKey, e.Method).
		Str(OwnerIDKey, e.OwnerID).
		Str(OwnerTypeKey, e.OwnerType).
		Str(PathKey, e.Path).
		Str(RequestIDKey, e.RequestID).
		Str(StatusTextKey, http.StatusText(e.StatusCode))

	if e.InnerError != nil {
		zle.AnErr(InternalErrorKey, e.InnerError)
	}
}

// FindOutermostAPIError returns the final APIError in the error chain.
func FindOutermostAPIError(err error) *APIError {
	res := FindAPIErrors(err)
	if len(res) > 0 {
		return res[0]
	}

	return nil
}

// FindAPIErrors returns a slice of all APIError found in the error chain.
func FindAPIErrors(err error) []*APIError {
	var errs []*APIError
	for {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			errs = append(errs, apiErr)
		}
		if err = errors.Unwrap(err); err == nil {
			break
		}
	}
	return errs
}

func GenerateRandomAPIError() APIError {
	return APIError{
		CallerID:    "caller-123",
		CallerType:  "admin",
		InnerError:  fmt.Errorf("wrapping error %w", errors.New("internal server error")),
		Method:      http.MethodGet,
		MultiParams: map[string][]string{"multiKey": {"multiVal1", "multiVal2"}},
		OwnerID:     "user-123",
		OwnerType:   "user",
		Path:        "/test/endpoint",
		PathParams:  map[string]string{"pathKey1": "pathVal1", "pathKey2": "pathVal2"},
		QueryParams: map[string]string{"queryKey1": "queryVal1", "queryKey2": "queryVal2"},
		RequestID:   "req-123",
	}
}

func GenerateNonRandomAPIError() APIError {
	return APIError{
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
}
