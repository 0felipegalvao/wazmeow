package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"wazmeow/internal/http/dto"
	"wazmeow/pkg/logger"
)

// ValidationMiddleware validates request content type and basic structure
func ValidationMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip validation for GET, DELETE, and OPTIONS requests
			if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check Content-Type for POST, PUT, PATCH requests only if there's a body
			if r.ContentLength > 0 {
				contentType := r.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					log.WarnWithFields("Invalid content type", logger.Fields{
						"content_type": contentType,
						"method":       r.Method,
						"path":         r.URL.Path,
					})

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnsupportedMediaType)

					response := dto.NewErrorResponse(
						"Unsupported media type. Expected application/json",
						"UNSUPPORTED_MEDIA_TYPE",
						"",
					)
					json.NewEncoder(w).Encode(response)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				// Generate a simple request ID (in production, use UUID)
				requestID = generateRequestID()
				r.Header.Set("X-Request-ID", requestID)
			}

			// Add request ID to response headers
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware recovers from panics and returns a proper error response
func RecoveryMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.ErrorWithFields("Panic recovered", logger.Fields{
						"error":      err,
						"method":     r.Method,
						"path":       r.URL.Path,
						"request_id": r.Header.Get("X-Request-ID"),
					})

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					response := dto.NewErrorResponse(
						"Internal server error",
						"INTERNAL_SERVER_ERROR",
						"An unexpected error occurred",
					)
					json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Remove server header
			w.Header().Del("Server")

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateJSONMiddleware validates that request body is valid JSON
func ValidateJSONMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip validation for GET, DELETE, and OPTIONS requests
			if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check if body is valid JSON
			if r.Body != nil && r.ContentLength > 0 {
				var temp interface{}
				decoder := json.NewDecoder(r.Body)
				if err := decoder.Decode(&temp); err != nil {
					log.WarnWithFields("Invalid JSON in request body", logger.Fields{
						"error":  err.Error(),
						"method": r.Method,
						"path":   r.URL.Path,
					})

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)

					response := dto.NewErrorResponse(
						"Invalid JSON in request body",
						"INVALID_JSON",
						err.Error(),
					)
					json.NewEncoder(w).Encode(response)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// In production, use a proper UUID library
	return "req-" + strings.ReplaceAll(strings.ReplaceAll(
		strings.ReplaceAll(
			"xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx",
			"x", "a"), // Simple replacement for demo
		"y", "b"), // Simple replacement for demo
		"-", "")[:16]
}
