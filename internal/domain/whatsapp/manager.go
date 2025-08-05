package whatsapp

import (
	"context"
	"errors"

	"wazmeow/internal/domain/session"
)

// ManagerConfig represents configuration for the WhatsApp manager
type ManagerConfig struct {
	DatabaseURL      string
	LogLevel         string
	QRTimeout        int
	ReconnectDelay   int
	MaxReconnects    int
	EnableWebhooks   bool
	WebhookURL       string
	EnableMetrics    bool
}

// ClientConfig represents configuration for a WhatsApp client
type ClientConfig struct {
	SessionID      session.SessionID
	DatabaseURL    string
	LogLevel       string
	QRTimeout      int
	ReconnectDelay int
	MaxReconnects  int
}

// ManagerStats represents statistics for the WhatsApp manager
type ManagerStats struct {
	TotalClients      int
	ConnectedClients  int
	AuthenticatedClients int
	ErrorClients      int
	Uptime           int64
}

// ClientStats represents statistics for a WhatsApp client
type ClientStats struct {
	SessionID        session.SessionID
	Status           ConnectionStatus
	JID              string
	ConnectedAt      int64
	AuthenticatedAt  int64
	MessagesSent     int64
	MessagesReceived int64
	LastActivity     int64
	Errors           int64
}

// WhatsApp domain errors
var (
	ErrClientNotFound      = errors.New("client not found")
	ErrClientAlreadyExists = errors.New("client already exists")
	ErrManagerNotRunning   = errors.New("manager not running")
	ErrInvalidConfig       = errors.New("invalid configuration")
	ErrConnectionFailed    = errors.New("connection failed")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrQRTimeout           = errors.New("QR code timeout")
	ErrInvalidPhoneNumber  = errors.New("invalid phone number")
	ErrMessageSendFailed   = errors.New("message send failed")
)

// AdvancedManager extends Manager with additional capabilities
type AdvancedManager interface {
	Manager

	// Statistics
	GetStats() *ManagerStats
	GetClientStats(sessionID session.SessionID) (*ClientStats, error)

	// Configuration
	UpdateConfig(config *ManagerConfig) error
	GetConfig() *ManagerConfig

	// Bulk operations
	ConnectAll(ctx context.Context) error
	DisconnectAll(ctx context.Context) error
	RestartClient(sessionID session.SessionID) error

	// Event handling
	SetGlobalEventHandler(handler EventHandler)
	RemoveGlobalEventHandler()

	// Health monitoring
	GetHealthStatus() map[session.SessionID]bool
	RestartUnhealthyClients(ctx context.Context) error
}

// Repository defines the interface for WhatsApp data persistence
type Repository interface {
	// Session data
	SaveSessionData(sessionID session.SessionID, data []byte) error
	LoadSessionData(sessionID session.SessionID) ([]byte, error)
	DeleteSessionData(sessionID session.SessionID) error

	// Device information
	SaveDeviceInfo(sessionID session.SessionID, info *DeviceInfo) error
	LoadDeviceInfo(sessionID session.SessionID) (*DeviceInfo, error)

	// Statistics
	SaveClientStats(sessionID session.SessionID, stats *ClientStats) error
	LoadClientStats(sessionID session.SessionID) (*ClientStats, error)

	// Cleanup
	CleanupOldData(maxAge int64) error
}
