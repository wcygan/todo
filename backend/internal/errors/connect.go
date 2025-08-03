package errors

import (
	"errors"

	"connectrpc.com/connect"
)

// ToConnectError converts an application error to a ConnectRPC error
func ToConnectError(err error) error {
	if err == nil {
		return nil
	}

	// If it's already a ConnectRPC error, return as-is
	if connect.CodeOf(err) != connect.CodeUnknown {
		return err
	}

	// Convert our custom errors to appropriate ConnectRPC codes
	var appErr *Error
	if !As(err, &appErr) {
		// If it's not our custom error, treat as internal error
		return connect.NewError(connect.CodeInternal, err)
	}

	switch appErr.Code {
	case CodeNotFound:
		return connect.NewError(connect.CodeNotFound, appErr)
	case CodeValidation:
		return connect.NewError(connect.CodeInvalidArgument, appErr)
	case CodeTimeout:
		return connect.NewError(connect.CodeDeadlineExceeded, appErr)
	case CodeInternal:
		return connect.NewError(connect.CodeInternal, appErr)
	default:
		return connect.NewError(connect.CodeInternal, appErr)
	}
}

// As is a convenience wrapper around errors.As for our Error type
func As(err error, target **Error) bool {
	return errors.As(err, target)
}