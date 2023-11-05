package sl

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

// apiError represents an error that occurred in the API layer of the application.
// It includes details like the HTTP status code and additional context.
type apiError struct {
	APIEndpoint   string `json:"apiEndpoint,omitempty"`
	CallerID      string `json:"callerId"`
	InternalError error  `json:"internalError,omitempty"` // An internal error if it exists such as twilio.SendSMS or other integrations
	Message       string `json:"message"`
	RequestID     string `json:"requestId"`
	StatusCode    int    `json:"statusCode"`
	StatusText    string `json:"statusText"`
	UserID        string `json:"userId"`

	execContext `json:"execContext"` // Embedded struct
}

// Error returns the string representation of the apiError.
func (e *apiError) Error() string {
	return fmt.Sprintf("[apiError] %d - %s at %s: %s", e.StatusCode, e.Message, e.APIEndpoint, e.InternalError)
}

// Unwrap provides the underlying error for use with errors.Is and errors.As functions.
func (e *apiError) Unwrap() error {
	return e.InternalError
}

// NewAPIErr is required because we have to json.Marshal apiError so execContext needs
// to be public however we don't want users to have to provide that
type NewAPIErr struct {
	APIEndpoint   string `json:"apiEndpoint,omitempty"`
	CallerID      string `json:"callerId"`
	InternalError error  `json:"internalError,omitempty"` // An internal error if it exists such as twilio.SendSMS or other integrations
	Message       string `json:"message"`
	RequestID     string `json:"requestId"`
	StatusCode    int    `json:"statusCode"`
	UserID        string `json:"userId"`
}

func LogNewAPIErr(newAPIErr NewAPIErr) error {
	if newAPIErr.Message == "" {
		newAPIErr.Message = "An API error occurred"
	}

	apiErr := apiError{
		APIEndpoint:   newAPIErr.APIEndpoint,
		CallerID:      newAPIErr.CallerID,
		InternalError: fmt.Errorf("wrapping error %w", newAPIErr.InternalError),
		Message:       newAPIErr.Message,
		RequestID:     newAPIErr.RequestID,
		StatusCode:    newAPIErr.StatusCode,
		StatusText:    http.StatusText(newAPIErr.StatusCode),
		UserID:        newAPIErr.UserID,

		execContext: getExecContext(),
	}

	log.Error().
		Object(ZLObjectKey, &apiErr).
		Msg(newAPIErr.Message)

	return &apiErr
}

// MarshalZerologObject allows apiError to be logged by zerolog.
func (e *apiError) MarshalZerologObject(zle *zerolog.Event) {
	zle.
		Int("line", e.Line).
		Int("statusCode", e.StatusCode).
		Str("apiEndpoint", e.APIEndpoint).
		Str("callerId", e.CallerID).
		Str("file", e.File).
		Str("function", e.Function).
		Str("message", e.Message).
		Str("requestId", e.RequestID).
		Str("userId", e.UserID)

	if e.InternalError != nil {
		zle.AnErr("internalError", e.InternalError)
	}
}
