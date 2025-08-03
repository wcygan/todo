package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents the type of error
type ErrorCode string

const (
	// CodeNotFound indicates a resource was not found
	CodeNotFound ErrorCode = "NOT_FOUND"
	// CodeValidation indicates invalid input data
	CodeValidation ErrorCode = "VALIDATION_ERROR"
	// CodeInternal indicates an internal server error
	CodeInternal ErrorCode = "INTERNAL_ERROR"
	// CodeTimeout indicates a request timeout
	CodeTimeout ErrorCode = "TIMEOUT"
)

// Error represents a structured application error
type Error struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
	Cause   error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements the errors.Unwrap interface for error chaining
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is implements the errors.Is interface for error comparison
func (e *Error) Is(target error) bool {
	var targetErr *Error
	if errors.As(target, &targetErr) {
		return e.Code == targetErr.Code
	}
	return false
}

// WithDetail adds a detail to the error
func (e *Error) WithDetail(key string, value interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// New creates a new Error with the given code and message
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   err,
	}
}

// NotFound creates a not found error
func NotFound(resource string, id string) *Error {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource)).
		WithDetail("resource", resource).
		WithDetail("id", id)
}

// Validation creates a validation error
func Validation(field string, reason string) *Error {
	return New(CodeValidation, fmt.Sprintf("validation failed for field '%s': %s", field, reason)).
		WithDetail("field", field).
		WithDetail("reason", reason)
}

// Internal creates an internal error
func Internal(message string) *Error {
	return New(CodeInternal, message)
}

// InternalWrap wraps an error as an internal error
func InternalWrap(err error, message string) *Error {
	return Wrap(err, CodeInternal, message)
}

// Timeout creates a timeout error
func Timeout(operation string) *Error {
	return New(CodeTimeout, fmt.Sprintf("operation '%s' timed out", operation)).
		WithDetail("operation", operation)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == CodeNotFound
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == CodeValidation
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == CodeInternal
}

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Code == CodeTimeout
}