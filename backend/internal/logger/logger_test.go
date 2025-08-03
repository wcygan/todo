package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wcygan/todo/backend/internal/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "json format",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "info",
					Format: "json",
				},
			},
		},
		{
			name: "text format",
			config: &config.Config{
				Logger: config.LoggerConfig{
					Level:  "debug",
					Format: "text",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)
			assert.NotNil(t, logger)
			assert.NotNil(t, logger.Logger)
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo}, // default
		{"", slog.LevelInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextOperations(t *testing.T) {
	ctx := context.Background()

	// Test adding and retrieving request ID
	requestID := "test-request-123"
	ctx = AddRequestIDToContext(ctx, requestID)
	
	retrievedID, ok := GetRequestIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, requestID, retrievedID)

	// Test adding and retrieving operation
	operation := "test-operation"
	ctx = AddOperationToContext(ctx, operation)
	
	retrievedOp, ok := GetOperationFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, operation, retrievedOp)

	// Test missing values
	emptyCtx := context.Background()
	_, ok = GetRequestIDFromContext(emptyCtx)
	assert.False(t, ok)
	
	_, ok = GetOperationFromContext(emptyCtx)
	assert.False(t, ok)
}

func TestLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	
	// Create logger with JSON handler for easier testing
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{Logger: slog.New(handler)}

	// Create context with request ID and operation
	ctx := context.Background()
	ctx = AddRequestIDToContext(ctx, "req-123")
	ctx = AddOperationToContext(ctx, "test-op")

	// Log a message
	logger.LogInfo(ctx, "test message", "key", "value")

	// Parse the JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Check that context values are included
	assert.Equal(t, "req-123", logEntry["request_id"])
	assert.Equal(t, "test-op", logEntry["operation"])
	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "value", logEntry["key"])
	assert.Equal(t, "INFO", logEntry["level"])
}

func TestLoggerWithError(t *testing.T) {
	var buf bytes.Buffer
	
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{Logger: slog.New(handler)}

	ctx := context.Background()
	testErr := assert.AnError

	// Log an error
	logger.LogError(ctx, "test error occurred", testErr, "additional", "info")

	// Parse the JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Check error is included
	assert.Equal(t, testErr.Error(), logEntry["error"])
	assert.Equal(t, "test error occurred", logEntry["msg"])
	assert.Equal(t, "info", logEntry["additional"])
	assert.Equal(t, "ERROR", logEntry["level"])
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{Logger: slog.New(handler)}
	ctx := context.Background()

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name:     "debug",
			logFunc:  func() { logger.LogDebug(ctx, "debug message") },
			expected: "DEBUG",
		},
		{
			name:     "info",
			logFunc:  func() { logger.LogInfo(ctx, "info message") },
			expected: "INFO",
		},
		{
			name:     "warn",
			logFunc:  func() { logger.LogWarn(ctx, "warn message") },
			expected: "WARN",
		},
		{
			name:     "error",
			logFunc:  func() { logger.LogError(ctx, "error message", assert.AnError) },
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestLoggerWithMethods(t *testing.T) {
	cfg := &config.Config{
		Logger: config.LoggerConfig{
			Level:  "debug",
			Format: "text",
		},
	}
	logger := New(cfg)

	// Test WithOperation
	opLogger := logger.WithOperation("test-operation")
	assert.NotNil(t, opLogger)

	// Test WithRequestID
	reqLogger := logger.WithRequestID("test-request-id")
	assert.NotNil(t, reqLogger)

	// Test WithError
	errLogger := logger.WithError(assert.AnError)
	assert.NotNil(t, errLogger)

	// Test method chaining
	chainedLogger := logger.WithOperation("op").WithRequestID("req").WithError(assert.AnError)
	assert.NotNil(t, chainedLogger)
}

func TestLoggerWithContextEmpty(t *testing.T) {
	cfg := &config.Config{
		Logger: config.LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}
	logger := New(cfg)

	// Test with empty context
	ctx := context.Background()
	contextLogger := logger.WithContext(ctx)
	assert.NotNil(t, contextLogger)

	// Test with context containing empty string values
	ctx = AddRequestIDToContext(ctx, "")
	ctx = AddOperationToContext(ctx, "")
	contextLogger = logger.WithContext(ctx)
	assert.NotNil(t, contextLogger)
}