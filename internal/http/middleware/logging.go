package middleware

import (
	"net/http"
	"time"

	"wazmeow/pkg/logger"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapper, r)

			// Log request
			duration := time.Since(start)
			
			fields := logger.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"query":       r.URL.RawQuery,
				"status_code": wrapper.statusCode,
				"duration_ms": duration.Milliseconds(),
				"user_agent":  r.UserAgent(),
				"remote_addr": r.RemoteAddr,
			}

			// Add request ID if present
			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				fields["request_id"] = requestID
			}

			// Log based on status code
			if wrapper.statusCode >= 500 {
				log.ErrorWithFields("HTTP request completed with server error", fields)
			} else if wrapper.statusCode >= 400 {
				log.WarnWithFields("HTTP request completed with client error", fields)
			} else {
				log.InfoWithFields("HTTP request completed", fields)
			}
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
