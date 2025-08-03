package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/wcygan/todo/backend/internal/config"
	"github.com/wcygan/todo/backend/internal/logger"
)

// timeoutResponseWriter wraps http.ResponseWriter to track if response was written
type timeoutResponseWriter struct {
	http.ResponseWriter
	written bool
}

func (w *timeoutResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *timeoutResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	return w.ResponseWriter.Write(data)
}

// TimeoutMiddleware adds request timeouts based on configuration
func TimeoutMiddleware(cfg *config.Config, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Server.ReadTimeout)
			defer cancel()

			// Add operation context for logging
			ctx = logger.AddOperationToContext(ctx, "http_request")

			// Create request with timeout context
			r = r.WithContext(ctx)

			// Wrap response writer to track writes
			tw := &timeoutResponseWriter{ResponseWriter: w}

			// Channel to signal completion
			done := make(chan struct{})
			
			// Execute request in goroutine
			go func() {
				defer close(done)
				next.ServeHTTP(tw, r)
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Request completed normally
				return
			case <-ctx.Done():
				// Request timed out
				log.LogWarn(ctx, "request timeout",
					"timeout", cfg.Server.ReadTimeout,
					"path", r.URL.Path,
					"method", r.Method,
				)
				
				// Only write timeout response if no response was written yet
				if !tw.written {
					w.WriteHeader(http.StatusRequestTimeout)
					w.Write([]byte(`{"error": "request timeout"}`))
				}
				return
			}
		})
	}
}

// ContextWithRequestTimeout creates a context with a timeout for individual operations
func ContextWithRequestTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// ContextWithDeadline creates a context with a deadline for batch operations
func ContextWithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, deadline)
}

// ContextWithCancel creates a cancellable context for long-running operations
func ContextWithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx)
}