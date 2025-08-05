package session

import (
	"net/url"
	"strings"
	"time"
)

// Session represents a WhatsApp session entity
type Session struct {
	id        SessionID
	name      string
	status    Status
	waJID     string
	qrCode    string
	proxyURL  string
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
}

// NewSession creates a new session with the given name
func NewSession(name string) *Session {
	if name == "" {
		panic("session name cannot be empty")
	}

	return &Session{
		id:        NewSessionID(),
		name:      name,
		status:    StatusDisconnected,
		waJID:     "",
		qrCode:    "",
		proxyURL:  "",
		isActive:  false,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}
}

// RestoreSession restores a session from persistence
func RestoreSession(id SessionID, name string, status Status, waJID string, qrCode string, proxyURL string, isActive bool, createdAt, updatedAt time.Time) *Session {
	return &Session{
		id:        id,
		name:      name,
		status:    status,
		waJID:     waJID,
		qrCode:    qrCode,
		proxyURL:  proxyURL,
		isActive:  isActive,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Connect marks the session as connected with the given WhatsApp JID
func (s *Session) Connect(waJID string) error {
	if s.status == StatusConnected {
		return ErrSessionAlreadyConnected
	}

	if waJID == "" {
		return ErrInvalidWhatsAppJID
	}

	s.waJID = waJID
	s.status = StatusConnected
	s.isActive = true
	s.updatedAt = time.Now()

	return nil
}

// Disconnect marks the session as disconnected
func (s *Session) Disconnect() {
	s.status = StatusDisconnected
	s.isActive = false
	s.updatedAt = time.Now()
}

// SetConnecting marks the session as connecting
func (s *Session) SetConnecting() {
	s.status = StatusConnecting
	s.updatedAt = time.Now()
}

// SetQRCode updates the session QR code
func (s *Session) SetQRCode(qrCode string) {
	s.qrCode = qrCode
	s.updatedAt = time.Now()
}

// ClearQRCode clears the session QR code
func (s *Session) ClearQRCode() {
	s.qrCode = ""
	s.updatedAt = time.Now()
}

// UpdateName updates the session name
func (s *Session) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidSessionName
	}

	s.name = name
	s.updatedAt = time.Now()
	return nil
}

// SetProxyURL updates the session proxy URL with validation
func (s *Session) SetProxyURL(proxyURL string) error {
	if proxyURL != "" {
		if err := s.validateProxyURL(proxyURL); err != nil {
			return err
		}
	}

	s.proxyURL = proxyURL
	s.updatedAt = time.Now()
	return nil
}

// ClearProxyURL clears the session proxy URL
func (s *Session) ClearProxyURL() {
	s.proxyURL = ""
	s.updatedAt = time.Now()
}

// HasProxy returns true if the session has a proxy configured
func (s *Session) HasProxy() bool {
	return s.proxyURL != ""
}

// GetProxyType returns the proxy type from the proxy URL
func (s *Session) GetProxyType() string {
	if !s.HasProxy() {
		return ""
	}

	if strings.HasPrefix(s.proxyURL, "http://") {
		return "http"
	} else if strings.HasPrefix(s.proxyURL, "https://") {
		return "https"
	} else if strings.HasPrefix(s.proxyURL, "socks4://") {
		return "socks4"
	} else if strings.HasPrefix(s.proxyURL, "socks5://") {
		return "socks5"
	}

	return "unknown"
}

// GetProxyHost returns the proxy host from the proxy URL
func (s *Session) GetProxyHost() string {
	if !s.HasProxy() {
		return ""
	}

	parsedURL, err := url.Parse(s.proxyURL)
	if err != nil {
		return ""
	}

	return parsedURL.Hostname()
}

// GetProxyPort returns the proxy port from the proxy URL
func (s *Session) GetProxyPort() string {
	if !s.HasProxy() {
		return ""
	}

	parsedURL, err := url.Parse(s.proxyURL)
	if err != nil {
		return ""
	}

	port := parsedURL.Port()
	if port == "" {
		// Return default ports based on scheme
		switch parsedURL.Scheme {
		case "http", "https":
			return "8080"
		case "socks4", "socks5":
			return "1080"
		}
	}

	return port
}

// HasProxyAuth returns true if the proxy URL contains authentication
func (s *Session) HasProxyAuth() bool {
	if !s.HasProxy() {
		return false
	}

	parsedURL, err := url.Parse(s.proxyURL)
	if err != nil {
		return false
	}

	return parsedURL.User != nil
}

// validateProxyURL validates the proxy URL format
func (s *Session) validateProxyURL(proxyURL string) error {
	if proxyURL == "" {
		return nil // Empty is valid (no proxy)
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return ErrInvalidProxyURL
	}

	// Check if scheme is supported
	supportedSchemes := []string{"http", "https", "socks4", "socks5"}
	schemeSupported := false
	for _, scheme := range supportedSchemes {
		if parsedURL.Scheme == scheme {
			schemeSupported = true
			break
		}
	}

	if !schemeSupported {
		return ErrUnsupportedProxyScheme
	}

	// Check if host is present
	if parsedURL.Hostname() == "" {
		return ErrInvalidProxyHost
	}

	return nil
}

// CanConnect returns true if the session can be connected
func (s *Session) CanConnect() bool {
	// Allow connection for disconnected and connecting states
	// Only prevent connection if already connected
	return s.status != StatusConnected
}

// IsConnected returns true if the session is connected
func (s *Session) IsConnected() bool {
	return s.status == StatusConnected && s.isActive
}

// IsConnecting returns true if the session is in connecting state
func (s *Session) IsConnecting() bool {
	return s.status == StatusConnecting
}

// Getters
func (s *Session) ID() SessionID {
	return s.id
}

func (s *Session) Name() string {
	return s.name
}

func (s *Session) Status() Status {
	return s.status
}

func (s *Session) WaJID() string {
	return s.waJID
}

func (s *Session) QRCode() string {
	return s.qrCode
}

func (s *Session) IsActive() bool {
	return s.isActive
}

func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Session) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Session) ProxyURL() string {
	return s.proxyURL
}

// Validate validates the session entity
func (s *Session) Validate() error {
	if s.name == "" {
		return ErrInvalidSessionName
	}

	if len(s.name) < 3 || len(s.name) > 50 {
		return ErrInvalidSessionName
	}

	return nil
}
