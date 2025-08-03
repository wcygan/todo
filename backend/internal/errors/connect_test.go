package errors

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
)

func TestToConnectError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode connect.Code
	}{
		{
			name:         "nil_error",
			err:          nil,
			expectedCode: connect.CodeUnknown, // nil case
		},
		{
			name:         "not_found_error",
			err:          NotFound("task", "123"),
			expectedCode: connect.CodeNotFound,
		},
		{
			name:         "validation_error",
			err:          Validation("field", "invalid"),
			expectedCode: connect.CodeInvalidArgument,
		},
		{
			name:         "timeout_error",
			err:          Timeout("operation"),
			expectedCode: connect.CodeDeadlineExceeded,
		},
		{
			name:         "internal_error",
			err:          Internal("internal error"),
			expectedCode: connect.CodeInternal,
		},
		{
			name:         "regular_error",
			err:          errors.New("regular error"),
			expectedCode: connect.CodeInternal,
		},
		{
			name:         "already_connect_error",
			err:          connect.NewError(connect.CodeUnauthenticated, errors.New("auth error")),
			expectedCode: connect.CodeUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToConnectError(tt.err)
			
			if tt.err == nil {
				assert.Nil(t, result)
				return
			}
			
			assert.Equal(t, tt.expectedCode, connect.CodeOf(result))
		})
	}
}

func TestAs(t *testing.T) {
	appErr := NotFound("task", "123")
	regularErr := errors.New("regular error")
	
	var target *Error
	
	// Should work with our custom error
	assert.True(t, As(appErr, &target))
	assert.Equal(t, CodeNotFound, target.Code)
	
	// Should not work with regular error
	target = nil
	assert.False(t, As(regularErr, &target))
	assert.Nil(t, target)
}