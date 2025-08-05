package whatsapp

import (
	"time"

	"wazmeow/internal/domain/session"
)

// Event represents a WhatsApp event
type Event struct {
	ID        string
	Type      EventType
	SessionID session.SessionID
	Timestamp time.Time
	Data      interface{}
}

// EventType represents the type of WhatsApp event
type EventType int

const (
	EventTypeConnected EventType = iota
	EventTypeDisconnected
	EventTypeQRCode
	EventTypeAuthenticated
	EventTypeAuthenticationFailed
	EventTypeMessage
	EventTypeMessageSent
	EventTypeMessageDelivered
	EventTypeMessageRead
	EventTypeError
	EventTypeReconnecting
	EventTypeContactUpdate
	EventTypeGroupUpdate
	EventTypePresenceUpdate
)

// String returns the string representation of EventType
func (t EventType) String() string {
	switch t {
	case EventTypeConnected:
		return "connected"
	case EventTypeDisconnected:
		return "disconnected"
	case EventTypeQRCode:
		return "qr_code"
	case EventTypeAuthenticated:
		return "authenticated"
	case EventTypeAuthenticationFailed:
		return "authentication_failed"
	case EventTypeMessage:
		return "message"
	case EventTypeMessageSent:
		return "message_sent"
	case EventTypeMessageDelivered:
		return "message_delivered"
	case EventTypeMessageRead:
		return "message_read"
	case EventTypeError:
		return "error"
	case EventTypeReconnecting:
		return "reconnecting"
	case EventTypeContactUpdate:
		return "contact_update"
	case EventTypeGroupUpdate:
		return "group_update"
	case EventTypePresenceUpdate:
		return "presence_update"
	default:
		return "unknown"
	}
}

// ConnectedEventData represents data for connection events
type ConnectedEventData struct {
	JID       string
	DeviceInfo *DeviceInfo
}

// DisconnectedEventData represents data for disconnection events
type DisconnectedEventData struct {
	Reason    string
	WillRetry bool
}

// QRCodeEventData represents data for QR code events
type QRCodeEventData struct {
	QRCode    string
	ExpiresAt time.Time
}

// AuthenticatedEventData represents data for authentication events
type AuthenticatedEventData struct {
	JID        string
	DeviceInfo *DeviceInfo
}

// AuthenticationFailedEventData represents data for authentication failure events
type AuthenticationFailedEventData struct {
	Reason string
	Fatal  bool
}

// MessageEventData represents data for message events
type MessageEventData struct {
	Message *Message
}

// ErrorEventData represents data for error events
type ErrorEventData struct {
	Error   error
	Code    string
	Fatal   bool
	Context map[string]interface{}
}

// ReconnectingEventData represents data for reconnection events
type ReconnectingEventData struct {
	Attempt   int
	MaxAttempts int
	Delay     time.Duration
}

// ContactUpdateEventData represents data for contact update events
type ContactUpdateEventData struct {
	JID    string
	Name   string
	Avatar string
}

// GroupUpdateEventData represents data for group update events
type GroupUpdateEventData struct {
	GroupJID    string
	Action      string
	ParticipantJID string
	AdminJID    string
}

// PresenceUpdateEventData represents data for presence update events
type PresenceUpdateEventData struct {
	JID       string
	Presence  string
	LastSeen  time.Time
}

// EventBus defines the interface for event publishing and subscription
type EventBus interface {
	// Publishing
	Publish(event *Event) error
	PublishAsync(event *Event)

	// Subscription
	Subscribe(eventType EventType, handler func(*Event)) error
	SubscribeToSession(sessionID session.SessionID, handler func(*Event)) error
	Unsubscribe(eventType EventType, handler func(*Event)) error
	UnsubscribeFromSession(sessionID session.SessionID, handler func(*Event)) error

	// Lifecycle
	Start() error
	Stop() error
	IsRunning() bool
}

// EventStore defines the interface for storing events
type EventStore interface {
	// Store events
	Store(event *Event) error
	StoreBatch(events []*Event) error

	// Retrieve events
	GetBySessionID(sessionID session.SessionID, limit, offset int) ([]*Event, error)
	GetByType(eventType EventType, limit, offset int) ([]*Event, error)
	GetByTimeRange(start, end time.Time, limit, offset int) ([]*Event, error)

	// Cleanup
	DeleteOldEvents(maxAge time.Duration) error
	DeleteBySessionID(sessionID session.SessionID) error

	// Statistics
	CountByType(eventType EventType) (int, error)
	CountBySessionID(sessionID session.SessionID) (int, error)
}

// WebhookHandler defines the interface for webhook handling
type WebhookHandler interface {
	// Send webhook
	SendWebhook(event *Event) error
	SendWebhookAsync(event *Event)

	// Configuration
	SetURL(url string)
	SetHeaders(headers map[string]string)
	SetTimeout(timeout time.Duration)
	SetRetryPolicy(maxRetries int, backoff time.Duration)

	// Health
	IsHealthy() bool
	GetStats() *WebhookStats
}

// WebhookStats represents webhook statistics
type WebhookStats struct {
	TotalSent     int64
	TotalFailed   int64
	AverageLatency time.Duration
	LastSentAt    time.Time
	LastError     error
}
