package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "simple_error",
			err:      New(CodeNotFound, "task not found"),
			expected: "NOT_FOUND: task not found",
		},
		{
			name:     "error_with_cause",
			err:      Wrap(errors.New("db connection failed"), CodeInternal, "database error"),
			expected: "INTERNAL_ERROR: database error (caused by: db connection failed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := Wrap(cause, CodeInternal, "wrapped error")
	
	assert.Equal(t, cause, errors.Unwrap(err))
}

func TestError_Is(t *testing.T) {
	err1 := New(CodeNotFound, "not found")
	err2 := New(CodeNotFound, "different message")
	err3 := New(CodeValidation, "validation error")
	
	assert.True(t, errors.Is(err1, err2))
	assert.False(t, errors.Is(err1, err3))
}

func TestError_WithDetail(t *testing.T) {
	err := New(CodeValidation, "invalid field").
		WithDetail("field", "email").
		WithDetail("value", "invalid-email")
	
	assert.Equal(t, "email", err.Details["field"])
	assert.Equal(t, "invalid-email", err.Details["value"])
}

func TestNotFound(t *testing.T) {
	err := NotFound("task", "123")
	
	assert.Equal(t, CodeNotFound, err.Code)
	assert.Contains(t, err.Message, "task not found")
	assert.Equal(t, "task", err.Details["resource"])
	assert.Equal(t, "123", err.Details["id"])
}

func TestValidation(t *testing.T) {
	err := Validation("email", "invalid format")
	
	assert.Equal(t, CodeValidation, err.Code)
	assert.Contains(t, err.Message, "email")
	assert.Contains(t, err.Message, "invalid format")
	assert.Equal(t, "email", err.Details["field"])
	assert.Equal(t, "invalid format", err.Details["reason"])
}

func TestInternal(t *testing.T) {
	err := Internal("database connection failed")
	
	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, "database connection failed", err.Message)
}

func TestInternalWrap(t *testing.T) {
	cause := errors.New("connection refused")
	err := InternalWrap(cause, "database error")
	
	assert.Equal(t, CodeInternal, err.Code)
	assert.Equal(t, "database error", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestTimeout(t *testing.T) {
	err := Timeout("create_task")
	
	assert.Equal(t, CodeTimeout, err.Code)
	assert.Contains(t, err.Message, "create_task")
	assert.Equal(t, "create_task", err.Details["operation"])
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is_not_found",
			err:      NotFound("task", "123"),
			expected: true,
		},
		{
			name:     "is_not_not_found",
			err:      Validation("field", "invalid"),
			expected: false,
		},
		{
			name:     "regular_error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNotFound(tt.err))
		})
	}
}

func TestIsValidation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is_validation",
			err:      Validation("field", "invalid"),
			expected: true,
		},
		{
			name:     "is_not_validation",
			err:      NotFound("task", "123"),
			expected: false,
		},
		{
			name:     "regular_error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidation(tt.err))
		})
	}
}

func TestIsInternal(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is_internal",
			err:      Internal("internal error"),
			expected: true,
		},
		{
			name:     "is_internal_wrapped",
			err:      InternalWrap(errors.New("cause"), "internal error"),
			expected: true,
		},
		{
			name:     "is_not_internal",
			err:      NotFound("task", "123"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsInternal(tt.err))
		})
	}
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is_timeout",
			err:      Timeout("operation"),
			expected: true,
		},
		{
			name:     "is_not_timeout",
			err:      NotFound("task", "123"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsTimeout(tt.err))
		})
	}
}