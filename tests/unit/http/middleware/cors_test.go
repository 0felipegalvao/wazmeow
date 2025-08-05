package http_middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"wazmeow/internal/http/middleware"
)

func TestCORSMiddleware(t *testing.T) {
	t.Run("should add CORS headers to response", func(t *testing.T) {
		// Arrange
		corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Wrap handler with middleware
		wrappedHandler := corsMiddleware(testHandler)

		// Create request
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test response", w.Body.String())

		// Check CORS headers
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin")) // No origin header in request
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials")) // Default is false
	})

	t.Run("should handle OPTIONS preflight request", func(t *testing.T) {
		// Arrange
		corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This should not be called for OPTIONS request
			t.Error("Handler should not be called for OPTIONS request")
		})

		// Wrap handler with middleware
		wrappedHandler := corsMiddleware(testHandler)

		// Create OPTIONS request
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, w.Code) // OPTIONS returns 204
		assert.Empty(t, w.Body.String())              // OPTIONS should return empty body

		// Check CORS headers
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin")) // Origin is set
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials")) // Default is false
	})

	t.Run("should preserve existing headers", func(t *testing.T) {
		// Arrange
		corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

		// Create a test handler that sets custom headers
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Custom-Header", "custom-value")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message": "created"}`))
		})

		// Wrap handler with middleware
		wrappedHandler := corsMiddleware(testHandler)

		// Create request
		req := httptest.NewRequest("POST", "/api/resource", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, `{"message": "created"}`, w.Body.String())

		// Check that custom headers are preserved
		assert.Equal(t, "custom-value", w.Header().Get("Custom-Header"))
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Check CORS headers are still added
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("should handle different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run("method_"+method, func(t *testing.T) {
				// Arrange
				corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

				// Create a test handler
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("method: " + r.Method))
				})

				// Wrap handler with middleware
				wrappedHandler := corsMiddleware(testHandler)

				// Create request
				req := httptest.NewRequest(method, "/test", nil)
				w := httptest.NewRecorder()

				// Act
				wrappedHandler.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "method: "+method, w.Body.String())

				// Check CORS headers are added for all methods
				assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
			})
		}
	})

	t.Run("should handle requests with different origins", func(t *testing.T) {
		origins := []string{
			"https://example.com",
			"http://localhost:3000",
			"https://app.domain.com",
			"", // No origin header
		}

		for _, origin := range origins {
			t.Run("origin_"+origin, func(t *testing.T) {
				// Arrange
				corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

				// Create a test handler
				testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})

				// Wrap handler with middleware
				wrappedHandler := corsMiddleware(testHandler)

				// Create request
				req := httptest.NewRequest("GET", "/test", nil)
				if origin != "" {
					req.Header.Set("Origin", origin)
				}
				w := httptest.NewRecorder()

				// Act
				wrappedHandler.ServeHTTP(w, req)

				// Assert
				assert.Equal(t, http.StatusOK, w.Code)

				// Check that Access-Control-Allow-Origin is always "*" (wildcard)
				assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
			})
		}
	})

	t.Run("should not interfere with error responses", func(t *testing.T) {
		// Arrange
		corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

		// Create a test handler that returns an error
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		})

		// Wrap handler with middleware
		wrappedHandler := corsMiddleware(testHandler)

		// Create request
		req := httptest.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal server error", w.Body.String())

		// Check CORS headers are still added even for error responses
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("should handle complex preflight request", func(t *testing.T) {
		// Arrange
		corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called for OPTIONS request")
		})

		// Wrap handler with middleware
		wrappedHandler := corsMiddleware(testHandler)

		// Create complex OPTIONS request
		req := httptest.NewRequest("OPTIONS", "/api/complex", nil)
		req.Header.Set("Origin", "https://frontend.example.com")
		req.Header.Set("Access-Control-Request-Method", "PUT")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization, X-Custom-Header")
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())

		// Check CORS headers
		assert.Equal(t, "https://frontend.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("should work with middleware chain", func(t *testing.T) {
		// Arrange
		corsMiddleware := middleware.CORSMiddleware(middleware.DefaultCORSConfig())

		// Create another middleware that adds a custom header
		customMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Custom-Middleware", "applied")
				next.ServeHTTP(w, r)
			})
		}

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("middleware chain"))
		})

		// Chain middlewares: CORS -> Custom -> Handler
		wrappedHandler := corsMiddleware(customMiddleware(testHandler))

		// Create request
		req := httptest.NewRequest("GET", "/chain", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "middleware chain", w.Body.String())

		// Check that both middlewares applied their headers
		assert.Equal(t, "applied", w.Header().Get("X-Custom-Middleware"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
	})
}
