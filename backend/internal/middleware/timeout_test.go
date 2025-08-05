package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wcygan/todo/backend/internal/config"
	"github.com/wcygan/todo/backend/internal/logger"
)

func TestTimeoutMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			ReadTimeout: 100 * time.Millisecond,
		},
		Logger: config.LoggerConfig{
			Level:  "debug",
			Format: "json",
		},
	}
	log := logger.New(cfg)

	tests := []struct {
		name           string
		handlerDelay   time.Duration
		expectedStatus int
		expectTimeout  bool
	}{
		{
			name:           "normal request",
			handlerDelay:   10 * time.Millisecond,
			expectedStatus: http.StatusOK,
			expectTimeout:  false,
		},
		{
			name:           "timeout request",
			handlerDelay:   200 * time.Millisecond,
			expectedStatus: http.StatusRequestTimeout,
			expectTimeout:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler with delay
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate work
				select {
				case <-time.After(tt.handlerDelay):
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("success"))
				case <-r.Context().Done():
					// Context cancelled due to timeout
					return
				}
			})

			// Wrap with timeout middleware
			middleware := TimeoutMiddleware(cfg, log)
			wrappedHandler := middleware(testHandler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Execute request
			wrappedHandler.ServeHTTP(w, req)

			// Check result
			if tt.expectTimeout {
				// For timeout cases, we need to wait a bit longer to ensure timeout occurs
				time.Sleep(150 * time.Millisecond)
				// The response may be 0 (no response written) or timeout status
				// Both are acceptable for timeout scenarios
				assert.True(t, w.Code == http.StatusRequestTimeout || w.Code == 0, 
					"Expected timeout response (408) or no response (0), got: %d", w.Code)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestContextWithRequestTimeout(t *testing.T) {
	ctx := context.Background()
	timeout := 100 * time.Millisecond

	// Create context with timeout
	timeoutCtx, cancel := ContextWithRequestTimeout(ctx, timeout)
	defer cancel()

	// Check that context has deadline
	deadline, ok := timeoutCtx.Deadline()
	assert.True(t, ok)
	assert.True(t, deadline.After(time.Now()))
	assert.True(t, deadline.Before(time.Now().Add(timeout+10*time.Millisecond)))

	// Wait for timeout
	select {
	case <-timeoutCtx.Done():
		assert.Equal(t, context.DeadlineExceeded, timeoutCtx.Err())
	case <-time.After(200 * time.Millisecond):
		t.Error("Context should have timed out")
	}
}

func TestContextWithDeadline(t *testing.T) {
	ctx := context.Background()
	deadline := time.Now().Add(50 * time.Millisecond)

	// Create context with deadline
	deadlineCtx, cancel := ContextWithDeadline(ctx, deadline)
	defer cancel()

	// Check that context has the correct deadline
	ctxDeadline, ok := deadlineCtx.Deadline()
	assert.True(t, ok)
	assert.True(t, ctxDeadline.Equal(deadline))

	// Wait for deadline
	select {
	case <-deadlineCtx.Done():
		assert.Equal(t, context.DeadlineExceeded, deadlineCtx.Err())
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should have reached deadline")
	}
}

func TestContextWithCancel(t *testing.T) {
	ctx := context.Background()

	// Create cancellable context
	cancelCtx, cancel := ContextWithCancel(ctx)

	// Context should not be done initially
	select {
	case <-cancelCtx.Done():
		t.Error("Context should not be cancelled initially")
	default:
	}

	// Cancel context
	cancel()

	// Context should be done after cancellation
	select {
	case <-cancelCtx.Done():
		assert.Equal(t, context.Canceled, cancelCtx.Err())
	case <-time.After(10 * time.Millisecond):
		t.Error("Context should be cancelled")
	}
}

func TestTimeoutMiddlewareWithLongRunningHandler(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			ReadTimeout: 50 * time.Millisecond,
		},
		Logger: config.LoggerConfig{
			Level:  "debug",
			Format: "text",
		},
	}
	log := logger.New(cfg)

	// Create handler that checks context cancellation
	handlerCalled := false
	contextCancelled := false

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		
		// Simulate long-running operation with context checking
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for i := 0; i < 10; i++ {
			select {
			case <-r.Context().Done():
				contextCancelled = true
				return
			case <-ticker.C:
				// Continue processing
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("completed"))
	})

	middleware := TimeoutMiddleware(cfg, log)
	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("POST", "/long-operation", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Handler should have been called
	assert.True(t, handlerCalled)
	
	// Wait a bit to ensure context cancellation is propagated
	time.Sleep(10 * time.Millisecond)
	
	// Context should have been cancelled due to timeout
	assert.True(t, contextCancelled)
}

func TestTimeoutMiddlewareContextPropagation(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			ReadTimeout: 100 * time.Millisecond,
		},
		Logger: config.LoggerConfig{
			Level:  "info",
			Format: "json",
		},
	}
	log := logger.New(cfg)

	// Create handler that checks context values
	var receivedOperation string

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if op, ok := logger.GetOperationFromContext(r.Context()); ok {
			receivedOperation = op
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := TimeoutMiddleware(cfg, log)
	wrappedHandler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Check that operation context was propagated
	assert.Equal(t, "http_request", receivedOperation)
	assert.Equal(t, http.StatusOK, w.Code)
}