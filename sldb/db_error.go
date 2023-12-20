package sldb

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seantcanavan/zerolog-json-structured-logs/slutil"
)

// DatabaseError represents an error that occurred in the database layer of the application.
// It includes details that might be relevant for debugging database issues.
type DatabaseError struct {
	Constraint string          `json:"constraint,omitempty"`
	DBName     string          `json:"dbName,omitempty"`
	InnerError error           `json:"innerError,omitempty"` // An inner error if it exists such as a SQL library Error
	Message    string          `json:"message,omitempty"`
	Operation  string          `json:"operation,omitempty"`
	Query      string          `json:"query,omitempty"`
	TableName  string          `json:"tableName,omitempty"`
	Type       EnumDBErrorType `json:"type,omitempty"`

	slutil.ExecContext `json:"execContext,omitempty"` // Embedded struct
}

// Error returns the string representation of the DatabaseError.
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("[DatabaseError] %s operation on %s.%s with query: %s - %s - %v",
		e.Operation, e.DBName, e.TableName, e.Query, e.Message, e.InnerError)
}

// Unwrap provides the underlying error for use with errors.Is and errors.As functions.
func (e *DatabaseError) Unwrap() error {
	return e.InnerError
}

// NewDBErr is required because we have to json.Marshal DatabaseError so execContext needs
// to be public however we don't want users to have to provide that
type NewDBErr struct {
	Constraint string          `json:"constraint,omitempty"`
	DBName     string          `json:"dbName,omitempty"`
	InnerError error           `json:"innerError,omitempty"` // An inner error if it exists such as a SQL library Error
	Message    string          `json:"message,omitempty"`
	Operation  string          `json:"operation,omitempty"`
	Query      string          `json:"query,omitempty"`
	TableName  string          `json:"tableName,omitempty"`
	Type       EnumDBErrorType `json:"type,omitempty"`
}

func LogNewDBErr(newDBErr NewDBErr) error {
	if newDBErr.Message == "" {
		newDBErr.Message = "A database error occurred"
	}

	dbErr := DatabaseError{
		Constraint: newDBErr.Constraint,
		DBName:     newDBErr.DBName,
		InnerError: fmt.Errorf("wrapping error %w", newDBErr.InnerError),
		Message:    newDBErr.Message,
		Operation:  newDBErr.Operation,
		Query:      newDBErr.Query,
		TableName:  newDBErr.TableName,
		Type:       newDBErr.Type,

		ExecContext: slutil.GetExecContext(),
	}

	log.Error().
		Object(slutil.ZLObjectKey, &dbErr).
		Msg(newDBErr.Message)

	return &dbErr
}

// MarshalZerologObject allows DatabaseError to be logged by zerolog.
func (e *DatabaseError) MarshalZerologObject(zle *zerolog.Event) {
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

	if e.InnerError != nil {
		zle.AnErr("innerError", e.InnerError)
	}
}

// FindOutermostDatabaseError returns the final APIError in the error chain.
func FindOutermostDatabaseError(err error) *DatabaseError {
	res := FindDatabaseErrors(err)
	if len(res) > 0 {
		return res[0]
	}

	return nil
}

// FindDatabaseErrors returns a slice of all APIError found in the error chain.
func FindDatabaseErrors(err error) []*DatabaseError {
	var errs []*DatabaseError
	for {
		var dbErr *DatabaseError
		if errors.As(err, &dbErr) {
			errs = append(errs, dbErr)
		}
		if err = errors.Unwrap(err); err == nil {
			break
		}
	}
	return errs
}
