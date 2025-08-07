package domain

import (
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
