package session

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// SessionID represents a unique session identifier
type SessionID struct {
	value string
}

// NewSessionID creates a new unique session ID
func NewSessionID() SessionID {
	return SessionID{value: uuid.New().String()}
}

// SessionIDFromString creates a SessionID from a string value
func SessionIDFromString(s string) (SessionID, error) {
	if s == "" {
		return SessionID{}, ErrInvalidSessionID
	}

	// Validate UUID format
	if _, err := uuid.Parse(s); err != nil {
		return SessionID{}, ErrInvalidSessionID
	}

	return SessionID{value: s}, nil
}

// String returns the string representation of the SessionID
func (id SessionID) String() string {
	return id.value
}

// IsEmpty returns true if the SessionID is empty
func (id SessionID) IsEmpty() bool {
	return id.value == ""
}

// Equals compares two SessionIDs for equality
func (id SessionID) Equals(other SessionID) bool {
	return id.value == other.value
}

// Status represents the connection status of a session
type Status int

const (
	// StatusDisconnected indicates the session is not connected
	StatusDisconnected Status = iota
	// StatusConnecting indicates the session is attempting to connect
	StatusConnecting
	// StatusConnected indicates the session is connected and active
	StatusConnected
)

// String returns the string representation of the Status
func (s Status) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	default:
		return "unknown"
	}
}

// IsValid returns true if the status is valid
func (s Status) IsValid() bool {
	return s >= StatusDisconnected && s <= StatusConnected
}

// StatusFromString creates a Status from a string value
func StatusFromString(s string) (Status, error) {
	switch strings.ToLower(s) {
	case "disconnected":
		return StatusDisconnected, nil
	case "connecting":
		return StatusConnecting, nil
	case "connected":
		return StatusConnected, nil
	default:
		return StatusDisconnected, fmt.Errorf("invalid status: %s", s)
	}
}

// SessionName represents a session name with validation
type SessionName struct {
	value string
}

// NewSessionName creates a new SessionName with validation
func NewSessionName(name string) (SessionName, error) {
	if err := validateSessionName(name); err != nil {
		return SessionName{}, err
	}

	return SessionName{value: name}, nil
}

// String returns the string representation of the SessionName
func (n SessionName) String() string {
	return n.value
}

// IsEmpty returns true if the SessionName is empty
func (n SessionName) IsEmpty() bool {
	return n.value == ""
}

// validateSessionName validates a session name
func validateSessionName(name string) error {
	if name == "" {
		return ErrInvalidSessionName
	}

	if len(name) < 3 {
		return ErrSessionNameTooShort
	}

	if len(name) > 50 {
		return ErrSessionNameTooLong
	}

	// Check for invalid characters (only alphanumeric, spaces, hyphens, underscores)
	for _, char := range name {
		if !isValidSessionNameChar(char) {
			return ErrInvalidSessionNameChars
		}
	}

	return nil
}

// isValidSessionNameChar checks if a character is valid for session names
func isValidSessionNameChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == ' ' ||
		char == '-' ||
		char == '_'
}

// SessionIdentifier represents a flexible session identifier that can be either a SessionID or SessionName
type SessionIdentifier struct {
	value          string
	identifierType IdentifierType
}

// IdentifierType represents the type of session identifier
type IdentifierType int

const (
	// IdentifierTypeID indicates the identifier is a SessionID (UUID)
	IdentifierTypeID IdentifierType = iota
	// IdentifierTypeName indicates the identifier is a SessionName
	IdentifierTypeName
)

// String returns the string representation of the IdentifierType
func (t IdentifierType) String() string {
	switch t {
	case IdentifierTypeID:
		return "id"
	case IdentifierTypeName:
		return "name"
	default:
		return "unknown"
	}
}

// NewSessionIdentifier creates a new SessionIdentifier with automatic type detection
func NewSessionIdentifier(value string) (SessionIdentifier, error) {
	if value == "" {
		return SessionIdentifier{}, ErrInvalidSessionIdentifier
	}

	// Trim whitespace to handle user input gracefully
	value = strings.TrimSpace(value)
	if value == "" {
		return SessionIdentifier{}, ErrInvalidSessionIdentifier
	}

	// Try to parse as UUID first (SessionID)
	if _, err := uuid.Parse(value); err == nil {
		return SessionIdentifier{
			value:          value,
			identifierType: IdentifierTypeID,
		}, nil
	}

	// If not a UUID, validate as SessionName
	if err := validateSessionName(value); err != nil {
		return SessionIdentifier{}, fmt.Errorf("invalid session identifier '%s': %w", value, err)
	}

	return SessionIdentifier{
		value:          value,
		identifierType: IdentifierTypeName,
	}, nil
}

// SessionIdentifierFromID creates a SessionIdentifier from a SessionID
func SessionIdentifierFromID(id SessionID) SessionIdentifier {
	return SessionIdentifier{
		value:          id.String(),
		identifierType: IdentifierTypeID,
	}
}

// SessionIdentifierFromName creates a SessionIdentifier from a SessionName
func SessionIdentifierFromName(name SessionName) SessionIdentifier {
	return SessionIdentifier{
		value:          name.String(),
		identifierType: IdentifierTypeName,
	}
}

// String returns the string representation of the SessionIdentifier
func (si SessionIdentifier) String() string {
	return si.value
}

// Type returns the type of the identifier
func (si SessionIdentifier) Type() IdentifierType {
	return si.identifierType
}

// IsID returns true if the identifier is a SessionID
func (si SessionIdentifier) IsID() bool {
	return si.identifierType == IdentifierTypeID
}

// IsName returns true if the identifier is a SessionName
func (si SessionIdentifier) IsName() bool {
	return si.identifierType == IdentifierTypeName
}

// ToSessionID converts the identifier to a SessionID if it's an ID type
func (si SessionIdentifier) ToSessionID() (SessionID, error) {
	if !si.IsID() {
		return SessionID{}, ErrInvalidSessionID
	}
	return SessionIDFromString(si.value)
}

// ToSessionName converts the identifier to a SessionName if it's a name type
func (si SessionIdentifier) ToSessionName() (SessionName, error) {
	if !si.IsName() {
		return SessionName{}, ErrInvalidSessionName
	}
	return NewSessionName(si.value)
}

// IsEmpty returns true if the SessionIdentifier is empty
func (si SessionIdentifier) IsEmpty() bool {
	return si.value == ""
}

// Equals compares two SessionIdentifiers for equality
func (si SessionIdentifier) Equals(other SessionIdentifier) bool {
	return si.value == other.value && si.identifierType == other.identifierType
}

// Validate validates the SessionIdentifier
func (si SessionIdentifier) Validate() error {
	if si.IsEmpty() {
		return ErrInvalidSessionIdentifier
	}

	if si.IsID() {
		// Validate UUID format
		if _, err := uuid.Parse(si.value); err != nil {
			return fmt.Errorf("invalid session ID format: %w", err)
		}
	} else if si.IsName() {
		// Validate session name
		if err := validateSessionName(si.value); err != nil {
			return fmt.Errorf("invalid session name: %w", err)
		}
	} else {
		return fmt.Errorf("unknown identifier type: %s", si.identifierType.String())
	}

	return nil
}

// WhatsAppJID represents a WhatsApp JID (Jabber ID)
type WhatsAppJID struct {
	value string
}

// NewWhatsAppJID creates a new WhatsAppJID with validation
func NewWhatsAppJID(jid string) (WhatsAppJID, error) {
	if jid == "" {
		return WhatsAppJID{}, ErrInvalidWhatsAppJID
	}

	// Basic JID validation (should contain @ symbol)
	if !strings.Contains(jid, "@") {
		return WhatsAppJID{}, ErrInvalidWhatsAppJID
	}

	return WhatsAppJID{value: jid}, nil
}

// String returns the string representation of the WhatsAppJID
func (j WhatsAppJID) String() string {
	return j.value
}

// IsEmpty returns true if the WhatsAppJID is empty
func (j WhatsAppJID) IsEmpty() bool {
	return j.value == ""
}

// Equals compares two WhatsAppJIDs for equality
func (j WhatsAppJID) Equals(other WhatsAppJID) bool {
	return j.value == other.value
}
