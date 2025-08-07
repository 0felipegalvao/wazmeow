package domain

import (
	"context"
)

// ConnectionStatus represents the WhatsApp connection status
type ConnectionStatus string

const (
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusConnecting   ConnectionStatus = "connecting"
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusFailed       ConnectionStatus = "failed"
	ConnectionStatusError        ConnectionStatus = "error"
)

// QRCodeEvent represents a QR code generation event
type QRCodeEvent struct {
	SessionID SessionID
	Code      string
	Timeout   int // seconds
}

// AuthenticationEvent represents an authentication event
type AuthenticationEvent struct {
	SessionID SessionID
	JID       string
	Success   bool
	Error     string
}

// ConnectionEvent represents a connection status change event
type ConnectionEvent struct {
	SessionID SessionID
	Status    ConnectionStatus
	Error     string
}

// Client defines the interface for WhatsApp client operations
type Client interface {
	// Connect establishes a connection for the given session
	Connect(ctx context.Context, sessionID SessionID) error

	// Disconnect closes the connection for the given session
	Disconnect(ctx context.Context, sessionID SessionID) error

	// Logout logs out and clears authentication for the session
	Logout(ctx context.Context, sessionID SessionID) error

	// GetQRCode generates a QR code for authentication
	GetQRCode(ctx context.Context, sessionID SessionID) (string, error)

	// PairPhone pairs a phone number for authentication
	PairPhone(ctx context.Context, sessionID SessionID, phone string) (string, error)

	// IsConnected checks if the session is connected
	IsConnected(ctx context.Context, sessionID SessionID) bool

	// IsAuthenticated checks if the session is authenticated
	IsAuthenticated(ctx context.Context, sessionID SessionID) bool

	// GetJID returns the WhatsApp JID for the session
	GetJID(ctx context.Context, sessionID SessionID) (string, error)

	// SetProxy configures proxy for the session
	SetProxy(ctx context.Context, sessionID SessionID, proxyURL string) error

	// GetConnectionStatus returns the current connection status
	GetConnectionStatus(ctx context.Context, sessionID SessionID) ConnectionStatus
}

// EventHandler defines the interface for handling WhatsApp events
type EventHandler interface {
	// OnQRCode is called when a QR code is generated
	OnQRCode(event QRCodeEvent)

	// OnAuthentication is called when authentication status changes
	OnAuthentication(event AuthenticationEvent)

	// OnConnection is called when connection status changes
	OnConnection(event ConnectionEvent)

	// OnMessage is called when a message is received
	OnMessage(event MessageEvent)

	// OnPresence is called when presence information is received
	OnPresence(event PresenceEvent)
}

// ClientManager manages multiple WhatsApp clients
type ClientManager interface {
	// CreateClient creates a new client for the session
	CreateClient(ctx context.Context, sessionID SessionID) (Client, error)

	// GetClient retrieves an existing client for the session
	GetClient(ctx context.Context, sessionID SessionID) (Client, error)

	// RemoveClient removes and cleans up a client for the session
	RemoveClient(ctx context.Context, sessionID SessionID) error

	// GetAllClients returns all active clients
	GetAllClients(ctx context.Context) map[SessionID]Client

	// SetEventHandler sets the event handler for all clients
	SetEventHandler(handler EventHandler)

	// ConnectAll connects all sessions marked as active
	ConnectAll(ctx context.Context) error

	// DisconnectAll disconnects all active sessions
	DisconnectAll(ctx context.Context) error
}

// ClientConfig represents configuration for WhatsApp clients
type ClientConfig struct {
	Debug      bool
	OSName     string
	ProxyURL   string
	Timeout    int // seconds
	RetryCount int
	LogLevel   string
}

// Factory creates WhatsApp clients with the given configuration
type Factory interface {
	// CreateClientManager creates a new client manager
	CreateClientManager(config ClientConfig) ClientManager

	// CreateClient creates a single client
	CreateClient(sessionID SessionID, config ClientConfig) (Client, error)
}
