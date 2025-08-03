package logger

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// RequestLoggingMiddleware adds request logging and request ID tracking
func RequestLoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Generate request ID
			requestID := generateRequestID()
			
			// Add request ID to context
			ctx := AddRequestIDToContext(r.Context(), requestID)
			r = r.WithContext(ctx)
			
			// Add request ID to response headers for debugging
			w.Header().Set("X-Request-ID", requestID)
			
			// Log incoming request
			logger.LogInfo(ctx, "incoming request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
			
			// Create response writer wrapper to capture status code
			wrappedWriter := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			// Process request
			next.ServeHTTP(wrappedWriter, r)
			
			// Log response
			duration := time.Since(start)
			logger.LogInfo(ctx, "request completed",
				"status_code", wrappedWriter.statusCode,
				"duration_ms", duration.Milliseconds(),
				"duration", duration.String(),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	bytes := make([]byte, 8) // 16 character hex string
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405")))
	}
	return hex.EncodeToString(bytes)
}