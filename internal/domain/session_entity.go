package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
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

// ParseSessionID parses a string into a SessionID
func ParseSessionID(s string) (SessionID, error) {
	if s == "" {
		return "", fmt.Errorf("session ID cannot be empty")
	}

	// Validate UUID format
	_, err := uuid.Parse(s)
	if err != nil {
		return "", fmt.Errorf("invalid session ID format: %w", err)
	}

	return SessionID(s), nil
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
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID              SessionID  `bun:",pk" json:"id"`
	Name            string     `bun:",notnull,unique" json:"name"`
	Status          Status     `bun:",default:'disconnected'" json:"status"`
	WebhookURL      string     `bun:"webhook_url" json:"webhook_url"`
	WAJID           string     `bun:"wa_jid" json:"wa_jid"`
	QRCode          string     `bun:"qr_code" json:"qr_code"`
	Events          string     `bun:",default:''" json:"events"`
	ProxyURL        string     `bun:"proxy_url" json:"proxy_url"`
	DeviceName      string     `bun:"device_name,default:'WazMeow'" json:"device_name"`
	IsActive        bool       `bun:"is_active,default:true" json:"is_active"`
	CreatedAt       time.Time  `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time  `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
	LastConnectedAt *time.Time `bun:"last_connected_at,nullzero" json:"last_connected_at,omitempty"`
}

// NewSession creates a new session with the given name
func NewSession(name string) *Session {
	now := time.Now()
	return &Session{
		ID:        NewSessionID(),
		Name:      strings.TrimSpace(name),
		Status:    StatusDisconnected,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// RestoreSession restores a session from stored data
func RestoreSession(
	id SessionID,
	name string,
	status Status,
	webhookURL string,
	wajid string,
	qrCode string,
	events string,
	proxyURL string,
	deviceName string,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	lastConnectedAt *time.Time,
) *Session {
	return &Session{
		ID:              id,
		Name:            name,
		Status:          status,
		WebhookURL:      webhookURL,
		WAJID:           wajid,
		QRCode:          qrCode,
		Events:          events,
		ProxyURL:        proxyURL,
		DeviceName:      deviceName,
		IsActive:        isActive,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		LastConnectedAt: lastConnectedAt,
	}
}

// Getters removed - using public fields directly

// Business methods
func (s *Session) UpdateName(name string) error {
	if name == "" {
		return NewValidationError("name cannot be empty")
	}
	s.Name = name
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Session) UpdateStatus(status Status) error {
	if !status.IsValid() {
		return NewValidationError("invalid status")
	}
	s.Status = status
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Session) SetWAJID(wajid string) {
	s.WAJID = wajid
	s.UpdatedAt = time.Now()
}

func (s *Session) SetQRCode(qrCode string) {
	s.QRCode = qrCode
	s.UpdatedAt = time.Now()
}

func (s *Session) SetProxyURL(proxyURL string) error {
	if proxyURL != "" {
		// Validate proxy URL format
		if _, err := url.Parse(proxyURL); err != nil {
			return NewValidationError("invalid proxy URL format")
		}
	}
	s.ProxyURL = proxyURL
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Session) Activate() {
	s.IsActive = true
	s.UpdatedAt = time.Now()
}

func (s *Session) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

func (s *Session) Connect() error {
	if s.Status == StatusConnected {
		return NewBusinessError("session is already connected")
	}
	s.Status = StatusConnecting
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Session) Disconnect() error {
	if s.Status == StatusDisconnected {
		return NewBusinessError("session is already disconnected")
	}
	s.Status = StatusDisconnected
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Session) MarkConnected() {
	s.Status = StatusConnected
	s.UpdatedAt = time.Now()
	now := time.Now()
	s.LastConnectedAt = &now
}

func (s *Session) MarkError() {
	s.Status = StatusError
	s.UpdatedAt = time.Now()
}

// CanConnect checks if the session can be connected
func (s *Session) CanConnect() bool {
	return s.Status == StatusDisconnected || s.Status == StatusError
}

// CanDisconnect checks if the session can be disconnected
func (s *Session) CanDisconnect() bool {
	return s.Status == StatusConnected || s.Status == StatusConnecting
}

// IsConnected checks if the session is connected
func (s *Session) IsConnected() bool {
	return s.Status == StatusConnected
}

// IsConnecting checks if the session is connecting
func (s *Session) IsConnecting() bool {
	return s.Status == StatusConnecting
}

// IsDisconnected checks if the session is disconnected
func (s *Session) IsDisconnected() bool {
	return s.Status == StatusDisconnected
}

// HasError checks if the session has an error
func (s *Session) HasError() bool {
	return s.Status == StatusError
}

// ToMap converts session to map for serialization
func (s *Session) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":                s.ID.String(),
		"name":              s.Name,
		"status":            string(s.Status),
		"webhook_url":       s.WebhookURL,
		"wa_jid":            s.WAJID,
		"qr_code":           s.QRCode,
		"events":            s.Events,
		"proxy_url":         s.ProxyURL,
		"device_name":       s.DeviceName,
		"is_active":         s.IsActive,
		"created_at":        s.CreatedAt,
		"updated_at":        s.UpdatedAt,
		"last_connected_at": s.LastConnectedAt,
	}
}
