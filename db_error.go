package sl

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// databaseError represents an error that occurred in the database layer of the application.
// It includes details that might be relevant for debugging database issues.
type databaseError struct {
	Constraint    string          `json:"constraint,omitempty"`
	DBName        string          `json:"dbName,omitempty"`
	InternalError error           `json:"internalError,omitempty"` // An internal error if it exists such as a SQL library Error
	Message       string          `json:"message,omitempty"`
	Operation     string          `json:"operation,omitempty"`
	Query         string          `json:"query,omitempty"`
	TableName     string          `json:"tableName,omitempty"`
	Type          EnumDBErrorType `json:"type,omitempty"`

	execContext `json:"execContext,omitempty"` // Embedded struct
}

// Error returns the string representation of the databaseError.
func (e *databaseError) Error() string {
	return fmt.Sprintf("[databaseError] %s operation on %s.%s with query: %s - %s - %v",
		e.Operation, e.DBName, e.TableName, e.Query, e.Message, e.InternalError)
}

// Unwrap provides the underlying error for use with errors.Is and errors.As functions.
func (e *databaseError) Unwrap() error {
	return e.InternalError
}

// NewDBErr is required because we have to json.Marshal databaseError so execContext needs
// to be public however we don't want users to have to provide that
type NewDBErr struct {
	Constraint    string          `json:"constraint,omitempty"`
	DBName        string          `json:"dbName,omitempty"`
	InternalError error           `json:"internalError,omitempty"` // An internal error if it exists such as a SQL library Error
	Message       string          `json:"message,omitempty"`
	Operation     string          `json:"operation,omitempty"`
	Query         string          `json:"query,omitempty"`
	TableName     string          `json:"tableName,omitempty"`
	Type          EnumDBErrorType `json:"type,omitempty"`
}

func LogNewDBErr(newDBErr NewDBErr) error {
	if newDBErr.Message == "" {
		newDBErr.Message = "A database error occurred"
	}

	dbErr := databaseError{
		Constraint:    newDBErr.Constraint,
		DBName:        newDBErr.DBName,
		InternalError: fmt.Errorf("wrapping error %w", newDBErr.InternalError),
		Message:       newDBErr.Message,
		Operation:     newDBErr.Operation,
		Query:         newDBErr.Query,
		TableName:     newDBErr.TableName,
		Type:          newDBErr.Type,

		execContext: getExecContext(),
	}

	log.Error().
		Object(ZLObjectKey, &dbErr).
		Msg(newDBErr.Message)

	return &dbErr
}

// MarshalZerologObject allows databaseError to be logged by zerolog.
func (e *databaseError) MarshalZerologObject(zle *zerolog.Event) {
	zle.
		Int("line", e.Line).
		Str("constraint", e.Constraint).
		Str("dbName", e.DBName).
		Str("file", e.File).
		Str("function", e.Function).
		Str("message", e.Message).
		Str("operation", e.Operation).
		Str("query", e.Query).
		Str("type", e.Type.String()).
		Str("tableName", e.TableName)

	if e.InternalError != nil {
		zle.AnErr("internalError", e.InternalError)
	}
}
