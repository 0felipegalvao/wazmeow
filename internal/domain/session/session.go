package domain

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SessionID represents a unique session identifier
type SessionID string

// NewSessionID generates a new unique session ID
func NewSessionID() SessionID {
	return SessionID(uuid.New().String())
}

// String returns the string representation of SessionID
func (id SessionID) String() string {
	return string(id)
}

// IsValid checks if the session ID is valid
func (id SessionID) IsValid() bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(string(id))
	return err == nil
}

// Status represents the session status
type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusError        Status = "error"
)

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusDisconnected, StatusConnecting, StatusConnected, StatusError:
		return true
	default:
		return false
	}
}

// Session represents a WhatsApp session
type Session struct {
	id        SessionID
	name      string
	status    Status
	wajid     string
	qrCode    string
	proxyURL  string
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
}

// NewSession creates a new session with the given name
func NewSession(name string) *Session {
	now := time.Now()
	return &Session{
		id:        NewSessionID(),
		name:      strings.TrimSpace(name),
		status:    StatusDisconnected,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}
}

// RestoreSession restores a session from stored data
func RestoreSession(
	id SessionID,
	name string,
	status Status,
	wajid string,
	qrCode string,
	proxyURL string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) *Session {
	return &Session{
		id:        id,
		name:      name,
		status:    status,
		wajid:     wajid,
		qrCode:    qrCode,
		proxyURL:  proxyURL,
		isActive:  isActive,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters
func (s *Session) ID() SessionID        { return s.id }
func (s *Session) Name() string         { return s.name }
func (s *Session) Status() Status       { return s.status }
func (s *Session) WAJID() string        { return s.wajid }
func (s *Session) QRCode() string       { return s.qrCode }
func (s *Session) ProxyURL() string     { return s.proxyURL }
func (s *Session) IsActive() bool       { return s.isActive }
func (s *Session) CreatedAt() time.Time { return s.createdAt }
func (s *Session) UpdatedAt() time.Time { return s.updatedAt }

// Business methods
func (s *Session) UpdateName(name string) error {
	if name == "" {
		return NewValidationError("name cannot be empty")
	}
	s.name = name
	s.updatedAt = time.Now()
	return nil
}

func (s *Session) UpdateStatus(status Status) error {
	if !status.IsValid() {
		return NewValidationError("invalid status")
	}
	s.status = status
	s.updatedAt = time.Now()
	return nil
}

func (s *Session) SetWAJID(wajid string) {
	s.wajid = wajid
	s.updatedAt = time.Now()
}

func (s *Session) SetQRCode(qrCode string) {
	s.qrCode = qrCode
	s.updatedAt = time.Now()
}

func (s *Session) SetProxyURL(proxyURL string) error {
	if proxyURL != "" {
		// Validate proxy URL format
		if _, err := url.Parse(proxyURL); err != nil {
			return NewValidationError("invalid proxy URL format")
		}
	}
	s.proxyURL = proxyURL
	s.updatedAt = time.Now()
	return nil
}

func (s *Session) Activate() {
	s.isActive = true
	s.updatedAt = time.Now()
}

func (s *Session) Deactivate() {
	s.isActive = false
	s.updatedAt = time.Now()
}

func (s *Session) Connect() error {
	if s.status == StatusConnected {
		return NewBusinessError("session is already connected")
	}
	s.status = StatusConnecting
	s.updatedAt = time.Now()
	return nil
}

func (s *Session) Disconnect() error {
	if s.status == StatusDisconnected {
		return NewBusinessError("session is already disconnected")
	}
	s.status = StatusDisconnected
	s.updatedAt = time.Now()
	return nil
}

func (s *Session) MarkConnected() {
	s.status = StatusConnected
	s.updatedAt = time.Now()
}

func (s *Session) MarkError() {
	s.status = StatusError
	s.updatedAt = time.Now()
}

// CanConnect checks if the session can be connected
func (s *Session) CanConnect() bool {
	return s.status == StatusDisconnected || s.status == StatusError
}

// CanDisconnect checks if the session can be disconnected
func (s *Session) CanDisconnect() bool {
	return s.status == StatusConnected || s.status == StatusConnecting
}

// IsConnected checks if the session is connected
func (s *Session) IsConnected() bool {
	return s.status == StatusConnected
}

// IsConnecting checks if the session is connecting
func (s *Session) IsConnecting() bool {
	return s.status == StatusConnecting
}

// IsDisconnected checks if the session is disconnected
func (s *Session) IsDisconnected() bool {
	return s.status == StatusDisconnected
}

// HasError checks if the session has an error
func (s *Session) HasError() bool {
	return s.status == StatusError
}

// ToMap converts session to map for serialization
func (s *Session) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         s.id.String(),
		"name":       s.name,
		"status":     string(s.status),
		"wa_jid":     s.wajid,
		"qr_code":    s.qrCode,
		"proxy_url":  s.proxyURL,
		"is_active":  s.isActive,
		"created_at": s.createdAt,
		"updated_at": s.updatedAt,
	}
}

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

// Repository defines the interface for session persistence
type Repository interface {
	// Create stores a new session
	Create(ctx context.Context, session *Session) error

	// GetByID retrieves a session by its ID
	GetByID(ctx context.Context, id SessionID) (*Session, error)

	// GetByName retrieves a session by its name
	GetByName(ctx context.Context, name string) (*Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *Session) error

	// Delete removes a session
	Delete(ctx context.Context, id SessionID) error

	// List retrieves all sessions with optional filters
	List(ctx context.Context, filters map[string]interface{}) ([]*Session, error)

	// ExistsByID checks if a session exists by ID
	ExistsByID(ctx context.Context, id SessionID) (bool, error)

	// ExistsByName checks if a session exists by name
	ExistsByName(ctx context.Context, name string) (bool, error)

	// GetActiveCount returns the count of active sessions
	GetActiveCount(ctx context.Context) (int64, error)

	// GetByStatus retrieves sessions by status
	GetByStatus(ctx context.Context, status Status) ([]*Session, error)

	// UpdateStatus updates the status of a session
	UpdateStatus(ctx context.Context, id SessionID, status Status) error

	// SetWAJID sets the WhatsApp JID for a session
	SetWAJID(ctx context.Context, id SessionID, wajid string) error

	// SetQRCode sets the QR code for a session
	SetQRCode(ctx context.Context, id SessionID, qrCode string) error

	// ClearQRCode clears the QR code for a session
	ClearQRCode(ctx context.Context, id SessionID) error

	// GetConnectedSessions retrieves all connected sessions
	GetConnectedSessions(ctx context.Context) ([]*Session, error)

	// BulkUpdateStatus updates status for multiple sessions
	BulkUpdateStatus(ctx context.Context, ids []SessionID, status Status) error
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
