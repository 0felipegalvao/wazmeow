package whatsapp

import (
	"context"
	"time"

	"wazmeow/internal/domain/session"
)

// Client defines the interface for WhatsApp client operations
type Client interface {
	// Connection management
	Connect(ctx context.Context) (*ConnectionResult, error)
	Disconnect(ctx context.Context) error
	IsConnected() bool
	GetConnectionStatus() ConnectionStatus

	// Authentication
	GenerateQR(ctx context.Context) (string, error)
	PairPhone(ctx context.Context, phoneNumber string) error
	IsAuthenticated() bool

	// Session information
	GetSessionID() session.SessionID
	GetJID() string
	GetDeviceInfo() *DeviceInfo

	// Messaging
	SendMessage(ctx context.Context, to, message string) error
	SendImage(ctx context.Context, to, imagePath, caption string) error
	SendDocument(ctx context.Context, to, documentPath, filename string) error

	// Event handling
	SetEventHandler(handler EventHandler)
	RemoveEventHandler()

	// Lifecycle
	Close() error
}

// Manager defines the interface for managing multiple WhatsApp clients
type Manager interface {
	// Client management
	CreateClient(sessionID session.SessionID) (Client, error)
	GetClient(sessionID session.SessionID) (Client, error)
	RemoveClient(sessionID session.SessionID) error
	ListClients() []session.SessionID

	// Lifecycle
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool

	// Health check
	HealthCheck() error
}

// ConnectionResult represents the result of a connection attempt
type ConnectionResult struct {
	JID       string
	QRCode    string
	Status    ConnectionStatus
	Error     error
	Timestamp time.Time
}

// ConnectionStatus represents the connection status
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusAuthenticating
	StatusAuthenticated
	StatusError
)

// String returns the string representation of ConnectionStatus
func (s ConnectionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusAuthenticating:
		return "authenticating"
	case StatusAuthenticated:
		return "authenticated"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// DeviceInfo represents device information
type DeviceInfo struct {
	Platform     string
	AppVersion   string
	DeviceModel  string
	OSVersion    string
	Manufacturer string
}

// EventHandler defines the interface for handling WhatsApp events
type EventHandler interface {
	OnConnected(sessionID session.SessionID, jid string)
	OnDisconnected(sessionID session.SessionID, reason string)
	OnQRCode(sessionID session.SessionID, qrCode string)
	OnAuthenticated(sessionID session.SessionID, jid string)
	OnAuthenticationFailed(sessionID session.SessionID, reason string)
	OnMessage(sessionID session.SessionID, message *Message)
	OnError(sessionID session.SessionID, err error)
}

// Message represents a WhatsApp message
type Message struct {
	ID        string
	From      string
	To        string
	Body      string
	Type      MessageType
	Timestamp time.Time
	IsFromMe  bool
}

// MessageType represents the type of message
type MessageType int

const (
	MessageTypeText MessageType = iota
	MessageTypeImage
	MessageTypeDocument
	MessageTypeAudio
	MessageTypeVideo
	MessageTypeSticker
	MessageTypeLocation
	MessageTypeContact
)

// String returns the string representation of MessageType
func (t MessageType) String() string {
	switch t {
	case MessageTypeText:
		return "text"
	case MessageTypeImage:
		return "image"
	case MessageTypeDocument:
		return "document"
	case MessageTypeAudio:
		return "audio"
	case MessageTypeVideo:
		return "video"
	case MessageTypeSticker:
		return "sticker"
	case MessageTypeLocation:
		return "location"
	case MessageTypeContact:
		return "contact"
	default:
		return "unknown"
	}
}
