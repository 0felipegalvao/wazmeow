package session

import (
	"errors"
	"fmt"
)

// Domain errors for session operations
var (
	// Session entity errors
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionAlreadyExists    = errors.New("session already exists")
	ErrSessionAlreadyConnected = errors.New("session already connected")
	ErrSessionNotConnected     = errors.New("session not connected")
	ErrSessionInvalidState     = errors.New("session in invalid state")

	// SessionID errors
	ErrInvalidSessionID = errors.New("invalid session ID")
	ErrEmptySessionID   = errors.New("session ID cannot be empty")

	// Session identifier errors
	ErrInvalidSessionIdentifier = errors.New("invalid session identifier")

	// Session name errors
	ErrInvalidSessionName      = errors.New("invalid session name")
	ErrSessionNameTooShort     = errors.New("session name too short (minimum 3 characters)")
	ErrSessionNameTooLong      = errors.New("session name too long (maximum 50 characters)")
	ErrInvalidSessionNameChars = errors.New("session name contains invalid characters")
	ErrSessionNameRequired     = errors.New("session name is required")

	// WhatsApp JID errors
	ErrInvalidWhatsAppJID = errors.New("invalid WhatsApp JID")
	ErrEmptyWhatsAppJID   = errors.New("WhatsApp JID cannot be empty")

	// Proxy URL errors
	ErrInvalidProxyURL        = errors.New("invalid proxy URL")
	ErrUnsupportedProxyScheme = errors.New("unsupported proxy scheme")
	ErrInvalidProxyHost       = errors.New("invalid proxy host")

	// Status errors
	ErrInvalidStatus = errors.New("invalid session status")

	// Repository errors
	ErrRepositoryConnection = errors.New("repository connection error")
	ErrRepositoryTimeout    = errors.New("repository operation timeout")
	ErrRepositoryConstraint = errors.New("repository constraint violation")

	// Validation errors
	ErrValidationFailed = errors.New("validation failed")
)

// SessionError represents a domain-specific error with additional context
type SessionError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *SessionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *SessionError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *SessionError) WithContext(key string, value interface{}) *SessionError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Error codes for different types of session errors
const (
	ErrCodeNotFound         = "SESSION_NOT_FOUND"
	ErrCodeAlreadyExists    = "SESSION_ALREADY_EXISTS"
	ErrCodeAlreadyConnected = "SESSION_ALREADY_CONNECTED"
	ErrCodeNotConnected     = "SESSION_NOT_CONNECTED"
	ErrCodeInvalidState     = "SESSION_INVALID_STATE"
	ErrCodeInvalidID        = "INVALID_SESSION_ID"
	ErrCodeInvalidName      = "INVALID_SESSION_NAME"
	ErrCodeInvalidJID       = "INVALID_WHATSAPP_JID"
	ErrCodeInvalidStatus    = "INVALID_STATUS"
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeRepository       = "REPOSITORY_ERROR"
)

// NewSessionError creates a new SessionError with the given code and message
func NewSessionError(code, message string) *SessionError {
	return &SessionError{
		Code:    code,
		Message: message,
	}
}

// NewSessionErrorWithCause creates a new SessionError with a cause
func NewSessionErrorWithCause(code, message string, cause error) *SessionError {
	return &SessionError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Convenience functions for creating common errors

// NewNotFoundError creates a session not found error
func NewNotFoundError(sessionID SessionID) *SessionError {
	return NewSessionError(ErrCodeNotFound, "session not found").
		WithContext("session_id", sessionID.String())
}

// NewAlreadyExistsError creates a session already exists error
func NewAlreadyExistsError(name string) *SessionError {
	return NewSessionError(ErrCodeAlreadyExists, "session already exists").
		WithContext("session_name", name)
}

// NewAlreadyConnectedError creates a session already connected error
func NewAlreadyConnectedError(sessionID SessionID) *SessionError {
	return NewSessionError(ErrCodeAlreadyConnected, "session already connected").
		WithContext("session_id", sessionID.String())
}

// NewInvalidNameError creates an invalid session name error
func NewInvalidNameError(name string, reason string) *SessionError {
	return NewSessionError(ErrCodeInvalidName, fmt.Sprintf("invalid session name: %s", reason)).
		WithContext("session_name", name).
		WithContext("reason", reason)
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *SessionError {
	return NewSessionError(ErrCodeValidation, fmt.Sprintf("validation failed for field '%s': %s", field, message)).
		WithContext("field", field).
		WithContext("validation_message", message)
}

// NewRepositoryError creates a repository error
func NewRepositoryError(operation string, cause error) *SessionError {
	return NewSessionErrorWithCause(ErrCodeRepository, fmt.Sprintf("repository operation failed: %s", operation), cause).
		WithContext("operation", operation)
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	var sessionErr *SessionError
	if errors.As(err, &sessionErr) {
		return sessionErr.Code == ErrCodeNotFound
	}
	return errors.Is(err, ErrSessionNotFound)
}

// IsAlreadyExistsError checks if the error is an already exists error
func IsAlreadyExistsError(err error) bool {
	var sessionErr *SessionError
	if errors.As(err, &sessionErr) {
		return sessionErr.Code == ErrCodeAlreadyExists
	}
	return errors.Is(err, ErrSessionAlreadyExists)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	var sessionErr *SessionError
	if errors.As(err, &sessionErr) {
		return sessionErr.Code == ErrCodeValidation
	}
	return errors.Is(err, ErrValidationFailed)
}

// IsRepositoryError checks if the error is a repository error
func IsRepositoryError(err error) bool {
	var sessionErr *SessionError
	if errors.As(err, &sessionErr) {
		return sessionErr.Code == ErrCodeRepository
	}
	return errors.Is(err, ErrRepositoryConnection) ||
		errors.Is(err, ErrRepositoryTimeout) ||
		errors.Is(err, ErrRepositoryConstraint)
}
