package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"wazmeow/internal/http/dto"
	"wazmeow/pkg/logger"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	APIKeys    []string
	SkipPaths  []string
	HeaderName string
}

// DefaultAuthConfig returns a default auth configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		APIKeys:    []string{}, // Empty means no auth required
		SkipPaths:  []string{"/health", "/metrics"},
		HeaderName: "X-API-Key",
	}
}

// AuthMiddleware implements API key authentication
func AuthMiddleware(config *AuthConfig, log logger.Logger) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			if shouldSkipAuth(r.URL.Path, config.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// If no API keys configured, skip auth
			if len(config.APIKeys) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Get API key from header
			apiKey := r.Header.Get(config.HeaderName)
			if apiKey == "" {
				// Also check Authorization header with Bearer prefix
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" {
				log.WarnWithFields("Missing API key", logger.Fields{
					"method":     r.Method,
					"path":       r.URL.Path,
					"remote_addr": r.RemoteAddr,
				})

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				
				response := dto.NewErrorResponse(
					"API key required",
					"UNAUTHORIZED",
					"Missing or invalid API key",
				)
				json.NewEncoder(w).Encode(response)
				return
			}

			// Validate API key
			if !isValidAPIKey(apiKey, config.APIKeys) {
				log.WarnWithFields("Invalid API key", logger.Fields{
					"method":     r.Method,
					"path":       r.URL.Path,
					"remote_addr": r.RemoteAddr,
					"api_key":    maskAPIKey(apiKey),
				})

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				
				response := dto.NewErrorResponse(
					"Invalid API key",
					"UNAUTHORIZED",
					"The provided API key is not valid",
				)
				json.NewEncoder(w).Encode(response)
				return
			}

			log.InfoWithFields("API key authenticated", logger.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"api_key":    maskAPIKey(apiKey),
			})

			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuthMiddleware implements HTTP Basic Authentication
func BasicAuthMiddleware(username, password string, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for health checks
			if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			// If no credentials configured, skip auth
			if username == "" || password == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Get credentials from request
			reqUsername, reqPassword, ok := r.BasicAuth()
			if !ok {
				log.WarnWithFields("Missing basic auth credentials", logger.Fields{
					"method":     r.Method,
					"path":       r.URL.Path,
					"remote_addr": r.RemoteAddr,
				})

				w.Header().Set("WWW-Authenticate", `Basic realm="WazMeow API"`)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				
				response := dto.NewErrorResponse(
					"Authentication required",
					"UNAUTHORIZED",
					"Basic authentication credentials required",
				)
				json.NewEncoder(w).Encode(response)
				return
			}

			// Validate credentials
			if reqUsername != username || reqPassword != password {
				log.WarnWithFields("Invalid basic auth credentials", logger.Fields{
					"method":     r.Method,
					"path":       r.URL.Path,
					"remote_addr": r.RemoteAddr,
					"username":   reqUsername,
				})

				w.Header().Set("WWW-Authenticate", `Basic realm="WazMeow API"`)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				
				response := dto.NewErrorResponse(
					"Invalid credentials",
					"UNAUTHORIZED",
					"The provided credentials are not valid",
				)
				json.NewEncoder(w).Encode(response)
				return
			}

			log.InfoWithFields("Basic auth authenticated", logger.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"username": reqUsername,
			})

			next.ServeHTTP(w, r)
		})
	}
}

// shouldSkipAuth checks if authentication should be skipped for a path
func shouldSkipAuth(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// isValidAPIKey checks if the provided API key is valid
func isValidAPIKey(apiKey string, validKeys []string) bool {
	for _, validKey := range validKeys {
		if apiKey == validKey {
			return true
		}
	}
	return false
}

// maskAPIKey masks an API key for logging
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}
