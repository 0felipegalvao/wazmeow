package handlers

import "time"

// SendTextMessageRequest represents a text message send request
type SendTextMessageRequest struct {
	Phone   string `json:"phone" validate:"required"`
	Message string `json:"message" validate:"required"`
	ID      string `json:"id,omitempty"`
}

// SendImageMessageRequest represents an image message send request
type SendImageMessageRequest struct {
	Phone   string `json:"phone" validate:"required"`
	Image   string `json:"image" validate:"required"` // Base64 or URL
	Caption string `json:"caption,omitempty"`
	ID      string `json:"id,omitempty"`
}

// SendAudioMessageRequest represents an audio message send request
type SendAudioMessageRequest struct {
	Phone string `json:"phone" validate:"required"`
	Audio string `json:"audio" validate:"required"` // Base64 or URL
	ID    string `json:"id,omitempty"`
}

// SendVideoMessageRequest represents a video message send request
type SendVideoMessageRequest struct {
	Phone   string `json:"phone" validate:"required"`
	Video   string `json:"video" validate:"required"` // Base64 or URL
	Caption string `json:"caption,omitempty"`
	ID      string `json:"id,omitempty"`
}

// SendDocumentMessageRequest represents a document message send request
type SendDocumentMessageRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Document string `json:"document" validate:"required"` // Base64 or URL
	Filename string `json:"filename,omitempty"`
	Mimetype string `json:"mimetype,omitempty"`
	ID       string `json:"id,omitempty"`
}

// SendLocationMessageRequest represents a location message send request
type SendLocationMessageRequest struct {
	Phone     string  `json:"phone" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
	ID        string  `json:"id,omitempty"`
}

// SendContactMessageRequest represents a contact message send request
type SendContactMessageRequest struct {
	Phone        string `json:"phone" validate:"required"`
	ContactPhone string `json:"contact_phone" validate:"required"`
	ContactName  string `json:"contact_name" validate:"required"`
	ID           string `json:"id,omitempty"`
}

// MessageResponse represents the response after sending a message
type MessageResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Phone     string    `json:"phone"`
	SessionID string    `json:"session_id"`
}
