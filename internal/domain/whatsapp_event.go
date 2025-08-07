package domain

import (
	"time"
)

// EventType represents the type of WhatsApp event
type EventType string

const (
	EventTypeMessage      EventType = "message"
	EventTypePresence     EventType = "presence"
	EventTypeReceipt      EventType = "receipt"
	EventTypeCall         EventType = "call"
	EventTypeGroup        EventType = "group"
	EventTypeContact      EventType = "contact"
	EventTypeStatus       EventType = "status"
	EventTypeNotification EventType = "notification"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVideo    MessageType = "video"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeDocument MessageType = "document"
	MessageTypeSticker  MessageType = "sticker"
	MessageTypeLocation MessageType = "location"
	MessageTypeContact  MessageType = "contact"
)

// PresenceType represents the presence status
type PresenceType string

const (
	PresenceTypeAvailable   PresenceType = "available"
	PresenceTypeUnavailable PresenceType = "unavailable"
	PresenceTypeComposing   PresenceType = "composing"
	PresenceTypeRecording   PresenceType = "recording"
	PresenceTypePaused      PresenceType = "paused"
)

// MessageEvent represents a message event
type MessageEvent struct {
	SessionID   SessionID      `json:"session_id"`
	EventType   EventType      `json:"event_type"`
	MessageID   string         `json:"message_id"`
	MessageType MessageType    `json:"message_type"`
	From        string         `json:"from"`
	To          string         `json:"to"`
	Timestamp   time.Time      `json:"timestamp"`
	Body        string         `json:"body,omitempty"`
	MediaURL    string         `json:"media_url,omitempty"`
	MimeType    string         `json:"mime_type,omitempty"`
	Caption     string         `json:"caption,omitempty"`
	IsGroup     bool           `json:"is_group"`
	GroupID     string         `json:"group_id,omitempty"`
	Participant string         `json:"participant,omitempty"`
	IsFromMe    bool           `json:"is_from_me"`
	Quoted      *QuotedMessage `json:"quoted,omitempty"`
	Mentions    []string       `json:"mentions,omitempty"`
	RawEvent    interface{}    `json:"raw_event,omitempty"`
}

// QuotedMessage represents a quoted message
type QuotedMessage struct {
	MessageID string      `json:"message_id"`
	From      string      `json:"from"`
	Body      string      `json:"body"`
	Type      MessageType `json:"type"`
}

// PresenceEvent represents a presence event
type PresenceEvent struct {
	SessionID SessionID    `json:"session_id"`
	EventType EventType    `json:"event_type"`
	From      string       `json:"from"`
	Presence  PresenceType `json:"presence"`
	Timestamp time.Time    `json:"timestamp"`
	IsGroup   bool         `json:"is_group"`
	GroupID   string       `json:"group_id,omitempty"`
}

// ReceiptEvent represents a message receipt event
type ReceiptEvent struct {
	SessionID SessionID `json:"session_id"`
	EventType EventType `json:"event_type"`
	MessageID string    `json:"message_id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "read", "delivered", "played"
}

// CallEvent represents a call event
type CallEvent struct {
	SessionID SessionID `json:"session_id"`
	EventType EventType `json:"event_type"`
	CallID    string    `json:"call_id"`
	From      string    `json:"from"`
	Timestamp time.Time `json:"timestamp"`
	IsVideo   bool      `json:"is_video"`
	Status    string    `json:"status"` // "offer", "accept", "reject", "timeout"
}

// GroupEvent represents a group event
type GroupEvent struct {
	SessionID   SessionID `json:"session_id"`
	EventType   EventType `json:"event_type"`
	GroupID     string    `json:"group_id"`
	Action      string    `json:"action"` // "create", "add", "remove", "promote", "demote", "subject", "description"
	Participant string    `json:"participant,omitempty"`
	Author      string    `json:"author"`
	Timestamp   time.Time `json:"timestamp"`
	Subject     string    `json:"subject,omitempty"`
	Description string    `json:"description,omitempty"`
}

// ContactEvent represents a contact event
type ContactEvent struct {
	SessionID SessionID `json:"session_id"`
	EventType EventType `json:"event_type"`
	JID       string    `json:"jid"`
	Name      string    `json:"name"`
	Notify    string    `json:"notify"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // "add", "update", "remove"
}

// StatusEvent represents a status event
type StatusEvent struct {
	SessionID SessionID `json:"session_id"`
	EventType EventType `json:"event_type"`
	From      string    `json:"from"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "text", "image", "video"
	Content   string    `json:"content,omitempty"`
	MediaURL  string    `json:"media_url,omitempty"`
}

// NotificationEvent represents a notification event
type NotificationEvent struct {
	SessionID SessionID   `json:"session_id"`
	EventType EventType   `json:"event_type"`
	Type      string      `json:"type"`
	From      string      `json:"from"`
	Timestamp time.Time   `json:"timestamp"`
	Content   interface{} `json:"content"`
}

// Event is a generic interface for all WhatsApp events
type Event interface {
	GetSessionID() SessionID
	GetEventType() EventType
	GetTimestamp() time.Time
}

// Implement Event interface for all event types
func (e MessageEvent) GetSessionID() SessionID { return e.SessionID }
func (e MessageEvent) GetEventType() EventType { return e.EventType }
func (e MessageEvent) GetTimestamp() time.Time { return e.Timestamp }

func (e PresenceEvent) GetSessionID() SessionID { return e.SessionID }
func (e PresenceEvent) GetEventType() EventType { return e.EventType }
func (e PresenceEvent) GetTimestamp() time.Time { return e.Timestamp }

func (e ReceiptEvent) GetSessionID() SessionID { return e.SessionID }
func (e ReceiptEvent) GetEventType() EventType { return e.EventType }
func (e ReceiptEvent) GetTimestamp() time.Time { return e.Timestamp }

func (e CallEvent) GetSessionID() SessionID { return e.SessionID }
func (e CallEvent) GetEventType() EventType { return e.EventType }
func (e CallEvent) GetTimestamp() time.Time { return e.Timestamp }

func (e GroupEvent) GetSessionID() SessionID { return e.SessionID }
func (e GroupEvent) GetEventType() EventType { return e.EventType }
func (e GroupEvent) GetTimestamp() time.Time { return e.Timestamp }

func (e ContactEvent) GetSessionID() SessionID { return e.SessionID }
func (e ContactEvent) GetEventType() EventType { return e.EventType }
func (e ContactEvent) GetTimestamp() time.Time { return e.Timestamp }

func (e StatusEvent) GetSessionID() SessionID { return e.SessionID }
func (e StatusEvent) GetEventType() EventType { return e.EventType }
func (e StatusEvent) GetTimestamp() time.Time { return e.Timestamp }

func (e NotificationEvent) GetSessionID() SessionID { return e.SessionID }
func (e NotificationEvent) GetEventType() EventType { return e.EventType }
func (e NotificationEvent) GetTimestamp() time.Time { return e.Timestamp }
