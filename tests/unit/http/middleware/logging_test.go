package http_middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"wazmeow/internal/http/middleware"
	"wazmeow/pkg/logger"
)

// MockLogger for middleware tests
type MockMiddlewareLogger struct {
	mock.Mock
}

func (m *MockMiddlewareLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockMiddlewareLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockMiddlewareLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockMiddlewareLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockMiddlewareLogger) DebugWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockMiddlewareLogger) InfoWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockMiddlewareLogger) WarnWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockMiddlewareLogger) ErrorWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockMiddlewareLogger) Fatal(msg string) {
	m.Called(msg)
}

func (m *MockMiddlewareLogger) FatalWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockMiddlewareLogger) DebugWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockMiddlewareLogger) InfoWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockMiddlewareLogger) WarnWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockMiddlewareLogger) ErrorWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockMiddlewareLogger) FatalWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockMiddlewareLogger) WithContext(ctx context.Context) logger.Logger {
	args := m.Called(ctx)
	return args.Get(0).(logger.Logger)
}

func (m *MockMiddlewareLogger) WithFields(fields logger.Fields) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockMiddlewareLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *MockMiddlewareLogger) WithError(err error) logger.Logger {
	args := m.Called(err)
	return args.Get(0).(logger.Logger)
}

func (m *MockMiddlewareLogger) SetLevel(level logger.Level) {
	m.Called(level)
}

func (m *MockMiddlewareLogger) GetLevel() logger.Level {
	args := m.Called()
	return args.Get(0).(logger.Level)
}

func (m *MockMiddlewareLogger) SetOutput(output io.Writer) {
	m.Called(output)
}

func (m *MockMiddlewareLogger) IsDebugEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMiddlewareLogger) IsInfoEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMiddlewareLogger) IsWarnEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMiddlewareLogger) IsErrorEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestLoggingMiddleware(t *testing.T) {
	t.Run("should log successful request", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockMiddlewareLogger)

		// Mock expectations - we expect InfoWithFields to be called with request details
		mockLogger.On("InfoWithFields", "HTTP request completed", mock.MatchedBy(func(fields logger.Fields) bool {
			// Verify that the fields contain expected keys
			_, hasMethod := fields["method"]
			_, hasPath := fields["path"]
			_, hasStatusCode := fields["status_code"]
			_, hasDuration := fields["duration_ms"]
			return hasMethod && hasPath && hasStatusCode && hasDuration
		})).Return()

		middleware := middleware.LoggingMiddleware(mockLogger)

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create request
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test response", w.Body.String())

		// Verify mocks
		mockLogger.AssertExpectations(t)
	})

	t.Run("should log request with error status", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockMiddlewareLogger)

		// Mock expectations
		mockLogger.On("ErrorWithFields", "HTTP request completed with server error", mock.MatchedBy(func(fields logger.Fields) bool {
			statusCode, ok := fields["status_code"]
			return ok && statusCode == 500
		})).Return()

		middleware := middleware.LoggingMiddleware(mockLogger)

		// Create a test handler that returns error
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create request
		req := httptest.NewRequest("POST", "/error", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal error", w.Body.String())

		// Verify mocks
		mockLogger.AssertExpectations(t)
	})

	t.Run("should log request details correctly", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockMiddlewareLogger)

		// Mock expectations with specific field validation
		mockLogger.On("InfoWithFields", "HTTP request completed", mock.MatchedBy(func(fields logger.Fields) bool {
			method, _ := fields["method"]
			path, _ := fields["path"]
			query, _ := fields["query"]
			statusCode, _ := fields["status_code"]
			userAgent, _ := fields["user_agent"]
			remoteAddr, _ := fields["remote_addr"]

			return method == "PUT" &&
				path == "/api/test" &&
				query == "param=value" &&
				statusCode == 201 &&
				userAgent == "custom-agent" &&
				remoteAddr != ""
		})).Return()

		middleware := middleware.LoggingMiddleware(mockLogger)

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created"))
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create request with specific details
		req := httptest.NewRequest("PUT", "/api/test?param=value", nil)
		req.Header.Set("User-Agent", "custom-agent")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "created", w.Body.String())

		// Verify mocks
		mockLogger.AssertExpectations(t)
	})

	t.Run("should measure request duration", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockMiddlewareLogger)

		// Mock expectations
		mockLogger.On("InfoWithFields", "HTTP request completed", mock.MatchedBy(func(fields logger.Fields) bool {
			duration, ok := fields["duration_ms"]
			if !ok {
				return false
			}

			// Duration should be a positive number
			durationMs, ok := duration.(int64)
			return ok && durationMs >= 0
		})).Return()

		middleware := middleware.LoggingMiddleware(mockLogger)

		// Create a test handler that takes some time
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Millisecond) // Small delay to ensure measurable duration
			w.WriteHeader(http.StatusOK)
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create request
		req := httptest.NewRequest("GET", "/slow", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify mocks
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle missing headers gracefully", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockMiddlewareLogger)

		// Mock expectations
		mockLogger.On("InfoWithFields", "HTTP request completed", mock.MatchedBy(func(fields logger.Fields) bool {
			userAgent, _ := fields["user_agent"]
			// User agent should be empty string when not provided
			return userAgent == ""
		})).Return()

		middleware := middleware.LoggingMiddleware(mockLogger)

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create request without User-Agent header
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify mocks
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle empty query string", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockMiddlewareLogger)

		// Mock expectations
		mockLogger.On("InfoWithFields", "HTTP request completed", mock.MatchedBy(func(fields logger.Fields) bool {
			query, _ := fields["query"]
			// Query should be empty string when not provided
			return query == ""
		})).Return()

		middleware := middleware.LoggingMiddleware(mockLogger)

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create request without query parameters
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify mocks
		mockLogger.AssertExpectations(t)
	})

	t.Run("should not interfere with response writing", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockMiddlewareLogger)

		// Mock expectations
		mockLogger.On("InfoWithFields", "HTTP request completed", mock.AnythingOfType("logger.Fields")).Return()

		middleware := middleware.LoggingMiddleware(mockLogger)

		// Create a test handler that writes custom headers and body
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Custom-Header", "custom-value")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("custom response body"))
		})

		// Wrap handler with middleware
		wrappedHandler := middleware(testHandler)

		// Create request
		req := httptest.NewRequest("POST", "/custom", nil)
		w := httptest.NewRecorder()

		// Act
		wrappedHandler.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, "custom-value", w.Header().Get("Custom-Header"))
		assert.Equal(t, "custom response body", w.Body.String())

		// Verify mocks
		mockLogger.AssertExpectations(t)
	})
}
