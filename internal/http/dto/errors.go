package dto

import (
	"errors"
	"fmt"
	"net/http"

	"wazmeow/internal/domain/session"
)

// ErrorCode represents standardized error codes for DTOs
type ErrorCode string

const (
	// Validation error codes
	ErrorCodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidInput      ErrorCode = "INVALID_INPUT"
	ErrorCodeMissingField      ErrorCode = "MISSING_FIELD"
	ErrorCodeInvalidFormat     ErrorCode = "INVALID_FORMAT"
	ErrorCodeInvalidLength     ErrorCode = "INVALID_LENGTH"
	ErrorCodeInvalidCharacters ErrorCode = "INVALID_CHARACTERS"

	// Session error codes
	ErrorCodeSessionNotFound      ErrorCode = "SESSION_NOT_FOUND"
	ErrorCodeSessionAlreadyExists ErrorCode = "SESSION_ALREADY_EXISTS"
	ErrorCodeSessionInvalidState  ErrorCode = "SESSION_INVALID_STATE"
	ErrorCodeSessionConnected     ErrorCode = "SESSION_ALREADY_CONNECTED"
	ErrorCodeSessionDisconnected  ErrorCode = "SESSION_DISCONNECTED"

	// Proxy error codes
	ErrorCodeInvalidProxy          ErrorCode = "INVALID_PROXY"
	ErrorCodeProxyConnectionFailed ErrorCode = "PROXY_CONNECTION_FAILED"
	ErrorCodeProxyAuthFailed       ErrorCode = "PROXY_AUTH_FAILED"

	// WhatsApp error codes
	ErrorCodeWhatsAppNotConnected ErrorCode = "WHATSAPP_NOT_CONNECTED"
	ErrorCodeWhatsAppAuthFailed   ErrorCode = "WHATSAPP_AUTH_FAILED"
	ErrorCodeWhatsAppQRExpired    ErrorCode = "WHATSAPP_QR_EXPIRED"

	// General error codes
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeTimeout            ErrorCode = "TIMEOUT"
	ErrorCodeRateLimited        ErrorCode = "RATE_LIMITED"
)

// String returns the string representation of ErrorCode
func (ec ErrorCode) String() string {
	return string(ec)
}

// HTTPStatusCode returns the appropriate HTTP status code for the error
func (ec ErrorCode) HTTPStatusCode() int {
	switch ec {
	case ErrorCodeValidationFailed, ErrorCodeInvalidInput, ErrorCodeMissingField,
		ErrorCodeInvalidFormat, ErrorCodeInvalidLength, ErrorCodeInvalidCharacters,
		ErrorCodeInvalidProxy:
		return http.StatusBadRequest
	case ErrorCodeSessionNotFound:
		return http.StatusNotFound
	case ErrorCodeSessionAlreadyExists:
		return http.StatusConflict
	case ErrorCodeSessionInvalidState, ErrorCodeSessionConnected, ErrorCodeSessionDisconnected,
		ErrorCodeWhatsAppNotConnected, ErrorCodeWhatsAppAuthFailed:
		return http.StatusUnprocessableEntity
	case ErrorCodeProxyConnectionFailed, ErrorCodeProxyAuthFailed:
		return http.StatusBadGateway
	case ErrorCodeWhatsAppQRExpired:
		return http.StatusGone
	case ErrorCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeTimeout:
		return http.StatusRequestTimeout
	case ErrorCodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// DTOError represents a structured error for DTOs
type DTOError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	StatusCode int                    `json:"-"`
}

// Error implements the error interface
func (de *DTOError) Error() string {
	if de.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", de.Code, de.Message, de.Details)
	}
	return fmt.Sprintf("%s: %s", de.Code, de.Message)
}

// NewDTOError creates a new DTO error
func NewDTOError(code ErrorCode, message string) *DTOError {
	return &DTOError{
		Code:       code,
		Message:    message,
		Context:    make(map[string]interface{}),
		StatusCode: code.HTTPStatusCode(),
	}
}

// WithDetails adds details to the error
func (de *DTOError) WithDetails(details string) *DTOError {
	de.Details = details
	return de
}

// WithContext adds context to the error
func (de *DTOError) WithContext(key string, value interface{}) *DTOError {
	if de.Context == nil {
		de.Context = make(map[string]interface{})
	}
	de.Context[key] = value
	return de
}

// WithStatusCode sets a custom status code
func (de *DTOError) WithStatusCode(statusCode int) *DTOError {
	de.StatusCode = statusCode
	return de
}

// ToErrorResponse converts the DTO error to an error response
func (de *DTOError) ToErrorResponse() *ErrorResponse {
	return NewErrorResponseWithContext(de.Message, de.Code.String(), de.Details, de.Context)
}

// ErrorMapper maps domain errors to DTO errors
type ErrorMapper struct{}

// NewErrorMapper creates a new error mapper
func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

// MapError maps a domain error to a DTO error
func (em *ErrorMapper) MapError(err error) *DTOError {
	if err == nil {
		return nil
	}

	// Handle validation errors
	if validationErr, ok := err.(ValidationError); ok {
		return NewDTOError(ErrorCodeValidationFailed, validationErr.Message).
			WithContext("field", validationErr.Field).
			WithContext("tag", validationErr.Tag).
			WithContext("value", validationErr.Value)
	}

	if validationErrs, ok := err.(ValidationErrors); ok {
		return NewDTOError(ErrorCodeValidationFailed, "Multiple validation errors").
			WithContext("errors", validationErrs)
	}

	// Handle session domain errors
	switch err {
	case session.ErrSessionNotFound:
		return NewDTOError(ErrorCodeSessionNotFound, "Session not found")
	case session.ErrSessionAlreadyExists:
		return NewDTOError(ErrorCodeSessionAlreadyExists, "Session already exists")
	case session.ErrSessionAlreadyConnected:
		return NewDTOError(ErrorCodeSessionConnected, "Session is already connected")
	case session.ErrInvalidSessionName:
		return NewDTOError(ErrorCodeInvalidInput, "Invalid session name")
	case session.ErrSessionNameTooShort:
		return NewDTOError(ErrorCodeInvalidLength, "Session name is too short")
	case session.ErrSessionNameTooLong:
		return NewDTOError(ErrorCodeInvalidLength, "Session name is too long")
	case session.ErrInvalidSessionNameChars:
		return NewDTOError(ErrorCodeInvalidCharacters, "Session name contains invalid characters")
	case session.ErrInvalidProxyURL:
		return NewDTOError(ErrorCodeInvalidProxy, "Invalid proxy URL")
	case session.ErrInvalidWhatsAppJID:
		return NewDTOError(ErrorCodeInvalidInput, "Invalid WhatsApp JID")
	}

	// Handle wrapped errors
	if wrappedErr := errors.Unwrap(err); wrappedErr != nil {
		if mappedErr := em.MapError(wrappedErr); mappedErr != nil {
			return mappedErr.WithDetails(err.Error())
		}
	}

	// Default to internal error
	return NewDTOError(ErrorCodeInternalError, "Internal server error").
		WithDetails(err.Error())
}

// MapErrorToResponse maps an error to an error response
func (em *ErrorMapper) MapErrorToResponse(err error) *ErrorResponse {
	dtoErr := em.MapError(err)
	return dtoErr.ToErrorResponse()
}

// ErrorResponseFactory creates standardized error responses
type ErrorResponseFactory struct {
	mapper *ErrorMapper
}

// NewErrorResponseFactory creates a new error response factory
func NewErrorResponseFactory() *ErrorResponseFactory {
	return &ErrorResponseFactory{
		mapper: NewErrorMapper(),
	}
}

// CreateValidationErrorResponse creates a validation error response
func (erf *ErrorResponseFactory) CreateValidationErrorResponse(errors []ValidationFieldError) *ValidationErrorResponse {
	return NewValidationErrorResponse(errors)
}

// CreateErrorResponse creates a generic error response
func (erf *ErrorResponseFactory) CreateErrorResponse(err error) *ErrorResponse {
	return erf.mapper.MapErrorToResponse(err)
}

// CreateNotFoundResponse creates a not found error response
func (erf *ErrorResponseFactory) CreateNotFoundResponse(resource string) *ErrorResponse {
	return NewDTOError(ErrorCodeSessionNotFound, fmt.Sprintf("%s not found", resource)).ToErrorResponse()
}

// CreateConflictResponse creates a conflict error response
func (erf *ErrorResponseFactory) CreateConflictResponse(resource string) *ErrorResponse {
	return NewDTOError(ErrorCodeSessionAlreadyExists, fmt.Sprintf("%s already exists", resource)).ToErrorResponse()
}

// CreateBadRequestResponse creates a bad request error response
func (erf *ErrorResponseFactory) CreateBadRequestResponse(message string) *ErrorResponse {
	return NewDTOError(ErrorCodeInvalidInput, message).ToErrorResponse()
}

// CreateInternalErrorResponse creates an internal error response
func (erf *ErrorResponseFactory) CreateInternalErrorResponse(details string) *ErrorResponse {
	return NewDTOError(ErrorCodeInternalError, "Internal server error").
		WithDetails(details).ToErrorResponse()
}

// CreateServiceUnavailableResponse creates a service unavailable error response
func (erf *ErrorResponseFactory) CreateServiceUnavailableResponse(service string) *ErrorResponse {
	return NewDTOError(ErrorCodeServiceUnavailable, fmt.Sprintf("%s service is unavailable", service)).ToErrorResponse()
}

// ErrorContext provides context for errors
type ErrorContext struct {
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	Timestamp  string                 `json:"timestamp,omitempty"`
	Additional map[string]interface{} `json:"additional,omitempty"`
}

// NewErrorContext creates a new error context
func NewErrorContext() *ErrorContext {
	return &ErrorContext{
		Additional: make(map[string]interface{}),
	}
}

// WithRequestID sets the request ID
func (ec *ErrorContext) WithRequestID(requestID string) *ErrorContext {
	ec.RequestID = requestID
	return ec
}

// WithUserID sets the user ID
func (ec *ErrorContext) WithUserID(userID string) *ErrorContext {
	ec.UserID = userID
	return ec
}

// WithSessionID sets the session ID
func (ec *ErrorContext) WithSessionID(sessionID string) *ErrorContext {
	ec.SessionID = sessionID
	return ec
}

// WithOperation sets the operation
func (ec *ErrorContext) WithOperation(operation string) *ErrorContext {
	ec.Operation = operation
	return ec
}

// WithTimestamp sets the timestamp
func (ec *ErrorContext) WithTimestamp(timestamp string) *ErrorContext {
	ec.Timestamp = timestamp
	return ec
}

// AddAdditional adds additional context
func (ec *ErrorContext) AddAdditional(key string, value interface{}) *ErrorContext {
	ec.Additional[key] = value
	return ec
}

// ToMap converts the error context to a map
func (ec *ErrorContext) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	if ec.RequestID != "" {
		result["request_id"] = ec.RequestID
	}
	if ec.UserID != "" {
		result["user_id"] = ec.UserID
	}
	if ec.SessionID != "" {
		result["session_id"] = ec.SessionID
	}
	if ec.Operation != "" {
		result["operation"] = ec.Operation
	}
	if ec.Timestamp != "" {
		result["timestamp"] = ec.Timestamp
	}

	for k, v := range ec.Additional {
		result[k] = v
	}

	return result
}
