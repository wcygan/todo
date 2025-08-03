package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/wcygan/todo/backend/internal/config"
)

// ContextKey is used for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request IDs
	RequestIDKey ContextKey = "request_id"
	// OperationKey is the context key for operation names
	OperationKey ContextKey = "operation"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// New creates a new logger based on the configuration
func New(cfg *config.Config) *Logger {
	var handler slog.Handler

	// Configure handler based on format
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Logger.Level),
	}

	if cfg.Logger.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithContext creates a new logger with context-specific fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	// Add request ID if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	// Add operation if present
	if operation, ok := ctx.Value(OperationKey).(string); ok && operation != "" {
		logger = logger.With("operation", operation)
	}

	return &Logger{Logger: logger}
}

// WithOperation creates a new logger with an operation field
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With("operation", operation),
	}
}

// WithRequestID creates a new logger with a request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", requestID),
	}
}

// WithError creates a new logger with error details
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
	}
}

// LogError logs an error with appropriate context
func (l *Logger) LogError(ctx context.Context, msg string, err error, args ...any) {
	logger := l.WithContext(ctx).WithError(err)
	logger.Error(msg, args...)
}

// LogInfo logs an info message with context
func (l *Logger) LogInfo(ctx context.Context, msg string, args ...any) {
	logger := l.WithContext(ctx)
	logger.Info(msg, args...)
}

// LogDebug logs a debug message with context
func (l *Logger) LogDebug(ctx context.Context, msg string, args ...any) {
	logger := l.WithContext(ctx)
	logger.Debug(msg, args...)
}

// LogWarn logs a warning message with context
func (l *Logger) LogWarn(ctx context.Context, msg string, args ...any) {
	logger := l.WithContext(ctx)
	logger.Warn(msg, args...)
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// AddRequestIDToContext adds a request ID to the context
func AddRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// AddOperationToContext adds an operation name to the context
func AddOperationToContext(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, OperationKey, operation)
}

// GetRequestIDFromContext retrieves the request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

// GetOperationFromContext retrieves the operation name from context
func GetOperationFromContext(ctx context.Context) (string, bool) {
	operation, ok := ctx.Value(OperationKey).(string)
	return operation, ok
}