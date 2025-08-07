package domain

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
	Type    string
	Message string
	Code    string
}

func (e DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// ValidationError represents a validation error
type ValidationError struct {
	DomainError
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		DomainError: DomainError{
			Type:    "VALIDATION_ERROR",
			Message: message,
			Code:    "VALIDATION_FAILED",
		},
	}
}

// BusinessError represents a business rule violation
type BusinessError struct {
	DomainError
}

// NewBusinessError creates a new business error
func NewBusinessError(message string) *BusinessError {
	return &BusinessError{
		DomainError: DomainError{
			Type:    "BUSINESS_ERROR",
			Message: message,
			Code:    "BUSINESS_RULE_VIOLATION",
		},
	}
}

// NotFoundError represents a not found error
type NotFoundError struct {
	DomainError
	Resource string
	ID       string
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		DomainError: DomainError{
			Type:    "NOT_FOUND_ERROR",
			Message: fmt.Sprintf("%s with ID '%s' not found", resource, id),
			Code:    "RESOURCE_NOT_FOUND",
		},
		Resource: resource,
		ID:       id,
	}
}

// AlreadyExistsError represents an already exists error
type AlreadyExistsError struct {
	DomainError
	Resource string
	Field    string
	Value    string
}

// NewAlreadyExistsError creates a new already exists error
func NewAlreadyExistsError(resource, field, value string) *AlreadyExistsError {
	return &AlreadyExistsError{
		DomainError: DomainError{
			Type:    "ALREADY_EXISTS_ERROR",
			Message: fmt.Sprintf("%s with %s '%s' already exists", resource, field, value),
			Code:    "RESOURCE_ALREADY_EXISTS",
		},
		Resource: resource,
		Field:    field,
		Value:    value,
	}
}

// Session-specific errors
func ErrSessionNotFound(id SessionID) error {
	return NewNotFoundError("Session", id.String())
}

func ErrSessionAlreadyExists(name string) error {
	return NewAlreadyExistsError("Session", "name", name)
}

func ErrInvalidSessionName(message string) error {
	return NewValidationError(fmt.Sprintf("invalid session name: %s", message))
}

func ErrCannotConnect(id SessionID, currentStatus Status) error {
	return NewBusinessError(fmt.Sprintf("cannot connect session %s: current status is %s", id, currentStatus))
}

func ErrCannotDisconnect(id SessionID, currentStatus Status) error {
	return NewBusinessError(fmt.Sprintf("cannot disconnect session %s: current status is %s", id, currentStatus))
}

func ErrInvalidStatus(status string) error {
	return NewValidationError(fmt.Sprintf("invalid session status: %s", status))
}
