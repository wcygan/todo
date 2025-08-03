package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wcygan/todo/backend/internal/config"
)

func TestRequestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	
	// Create logger with JSON handler for easier testing
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{Logger: slog.New(handler)}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request ID is in context
		requestID, ok := GetRequestIDFromContext(r.Context())
		assert.True(t, ok)
		assert.NotEmpty(t, requestID)

		// Write a response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with middleware
	middleware := RequestLoggingMiddleware(logger)
	wrappedHandler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response headers
	assert.Equal(t, http.StatusOK, w.Code)
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)

	// Parse log entries
	logLines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, logLines, 2) // Should have incoming request and request completed logs

	// Parse first log entry (incoming request)
	var incomingLog map[string]interface{}
	err := json.Unmarshal([]byte(logLines[0]), &incomingLog)
	require.NoError(t, err)

	assert.Equal(t, "incoming request", incomingLog["msg"])
	assert.Equal(t, "GET", incomingLog["method"])
	assert.Equal(t, "/test", incomingLog["path"])
	assert.Equal(t, "test-agent", incomingLog["user_agent"])
	assert.Equal(t, requestID, incomingLog["request_id"])

	// Parse second log entry (request completed)
	var completedLog map[string]interface{}
	err = json.Unmarshal([]byte(logLines[1]), &completedLog)
	require.NoError(t, err)

	assert.Equal(t, "request completed", completedLog["msg"])
	assert.Equal(t, float64(200), completedLog["status_code"]) // JSON numbers are float64
	assert.Contains(t, completedLog, "duration_ms")
	assert.Contains(t, completedLog, "duration")
	assert.Equal(t, requestID, completedLog["request_id"])
}

func TestRequestLoggingMiddlewareErrorResponse(t *testing.T) {
	var buf bytes.Buffer
	
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{Logger: slog.New(handler)}

	// Create a test handler that returns an error
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error occurred"))
	})

	// Wrap with middleware
	middleware := RequestLoggingMiddleware(logger)
	wrappedHandler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("POST", "/error", nil)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Parse log entries
	logLines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, logLines, 2)

	// Parse completed log entry
	var completedLog map[string]interface{}
	err := json.Unmarshal([]byte(logLines[1]), &completedLog)
	require.NoError(t, err)

	assert.Equal(t, "request completed", completedLog["msg"])
	assert.Equal(t, float64(500), completedLog["status_code"])
}

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// Test default status code
	assert.Equal(t, http.StatusOK, rw.statusCode)

	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, rw.statusCode)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test Write (should not change status code if already set)
	rw.Write([]byte("test"))
	assert.Equal(t, http.StatusNotFound, rw.statusCode)
}

func TestGenerateRequestID(t *testing.T) {
	// Generate multiple request IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()
		assert.NotEmpty(t, id)
		assert.Equal(t, 16, len(id)) // 8 bytes = 16 hex characters
		
		// Check uniqueness
		assert.False(t, ids[id], "Request ID should be unique: %s", id)
		ids[id] = true

		// Check it's valid hex
		assert.Regexp(t, "^[0-9a-f]+$", id)
	}
}

func TestMiddlewareIntegration(t *testing.T) {
	cfg := &config.Config{
		Logger: config.LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}
	logger := New(cfg)

	// Create a simple handler that uses the logger
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log something from within the handler
		logger.LogInfo(r.Context(), "handler executed", "endpoint", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	middleware := RequestLoggingMiddleware(logger)
	wrappedHandler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check that request completed successfully
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestMiddlewareWithComplexPath(t *testing.T) {
	var buf bytes.Buffer
	
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{Logger: slog.New(handler)}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	middleware := RequestLoggingMiddleware(logger)
	wrappedHandler := middleware(testHandler)

	// Test with complex path and query parameters
	req := httptest.NewRequest("POST", "/api/v1/tasks?filter=active&sort=created", strings.NewReader(`{"description":"test task"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Parse log entries
	logLines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, logLines, 2)

	// Check incoming request log
	var incomingLog map[string]interface{}
	err := json.Unmarshal([]byte(logLines[0]), &incomingLog)
	require.NoError(t, err)

	assert.Equal(t, "POST", incomingLog["method"])
	assert.Equal(t, "/api/v1/tasks", incomingLog["path"]) // Query params not included in path
	assert.Equal(t, "Mozilla/5.0", incomingLog["user_agent"])

	// Check completed request log
	var completedLog map[string]interface{}
	err = json.Unmarshal([]byte(logLines[1]), &completedLog)
	require.NoError(t, err)

	assert.Equal(t, float64(201), completedLog["status_code"])
}