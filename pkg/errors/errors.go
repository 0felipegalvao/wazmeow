package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound represents not found errors
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeConflict represents conflict errors
	ErrorTypeConflict ErrorType = "conflict"
	// ErrorTypeUnauthorized represents unauthorized errors
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	// ErrorTypeForbidden represents forbidden errors
	ErrorTypeForbidden ErrorType = "forbidden"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeBadRequest represents bad request errors
	ErrorTypeBadRequest ErrorType = "bad_request"
	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeRateLimit represents rate limit errors
	ErrorTypeRateLimit ErrorType = "rate_limit"
)

// AppError represents an application error with additional context
type AppError struct {
	Type       ErrorType              `json:"type"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Cause      error                  `json:"-"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	HTTPStatus int                    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithStackTrace adds stack trace to the error
func (e *AppError) WithStackTrace() *AppError {
	e.StackTrace = getStackTrace()
	return e
}

// ToJSON converts the error to JSON
func (e *AppError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// GetHTTPStatus returns the HTTP status code for the error
func (e *AppError) GetHTTPStatus() int {
	if e.HTTPStatus != 0 {
		return e.HTTPStatus
	}

	switch e.Type {
	case ErrorTypeValidation, ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithCause creates a new application error with a cause
func NewAppErrorWithCause(errorType ErrorType, code, message string, cause error) *AppError {
	return &AppError{
		Type:    errorType,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Convenience functions for creating common errors

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return NewAppError(ErrorTypeValidation, "VALIDATION_ERROR", message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrorTypeNotFound, "NOT_FOUND", fmt.Sprintf("%s not found", resource))
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return NewAppError(ErrorTypeConflict, "CONFLICT", message)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "unauthorized"
	}
	return NewAppError(ErrorTypeUnauthorized, "UNAUTHORIZED", message)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "forbidden"
	}
	return NewAppError(ErrorTypeForbidden, "FORBIDDEN", message)
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *AppError {
	if message == "" {
		message = "internal server error"
	}
	return NewAppError(ErrorTypeInternal, "INTERNAL_ERROR", message).WithStackTrace()
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return NewAppError(ErrorTypeBadRequest, "BAD_REQUEST", message)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string) *AppError {
	if message == "" {
		message = "request timeout"
	}
	return NewAppError(ErrorTypeTimeout, "TIMEOUT", message)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *AppError {
	if message == "" {
		message = "rate limit exceeded"
	}
	return NewAppError(ErrorTypeRateLimit, "RATE_LIMIT", message)
}

// Wrap wraps an existing error as an AppError
func Wrap(err error, errorType ErrorType, code, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return it
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewAppErrorWithCause(errorType, code, message, err)
}

// WrapInternal wraps an error as an internal error
func WrapInternal(err error, message string) *AppError {
	if err == nil {
		return nil
	}
	return Wrap(err, ErrorTypeInternal, "INTERNAL_ERROR", message).WithStackTrace()
}

// IsType checks if an error is of a specific type
func IsType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return IsType(err, ErrorTypeValidation)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return IsType(err, ErrorTypeNotFound)
}

// IsConflictError checks if an error is a conflict error
func IsConflictError(err error) bool {
	return IsType(err, ErrorTypeConflict)
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return IsType(err, ErrorTypeUnauthorized)
}

// IsForbiddenError checks if an error is a forbidden error
func IsForbiddenError(err error) bool {
	return IsType(err, ErrorTypeForbidden)
}

// IsInternalError checks if an error is an internal error
func IsInternalError(err error) bool {
	return IsType(err, ErrorTypeInternal)
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var traces []string
	for {
		frame, more := frames.Next()
		traces = append(traces, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return strings.Join(traces, "\n")
}

// ErrorResponse represents an HTTP error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details string                 `json:"details,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ToErrorResponse converts an AppError to an ErrorResponse
func ToErrorResponse(err error) *ErrorResponse {
	if appErr, ok := err.(*AppError); ok {
		return &ErrorResponse{
			Error:   appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
			Context: appErr.Context,
		}
	}

	return &ErrorResponse{
		Error: err.Error(),
	}
}
