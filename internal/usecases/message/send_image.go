package message

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// SendImageMessageUseCase handles sending WhatsApp image messages
type SendImageMessageUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
	validator   validator.Validator
}

// NewSendImageMessageUseCase creates a new send image message use case
func NewSendImageMessageUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger, validator validator.Validator) *SendImageMessageUseCase {
	return &SendImageMessageUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
		validator:   validator,
	}
}

// SendImageMessageRequest represents the request to send an image message
type SendImageMessageRequest struct {
	SessionID   session.SessionID `json:"session_id"`
	To          string            `json:"to" validate:"required"`
	Image       string            `json:"image" validate:"required"` // Base64 string
	Caption     string            `json:"caption" validate:"max=1024"`
	MimeType    string            `json:"mime_type"`
	ContextInfo interface{}       `json:"context_info,omitempty"`
}

// SendImageMessageResponse represents the response from sending an image message
type SendImageMessageResponse struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to"`
	Success   bool              `json:"success"`
	MessageID string            `json:"message_id,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// Execute sends an image message via WhatsApp
func (uc *SendImageMessageUseCase) Execute(ctx context.Context, req SendImageMessageRequest) (*SendImageMessageResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for send image message", err, logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, err
	}

	// Additional validation for image data
	if strings.TrimSpace(req.Image) == "" {
		uc.logger.WarnWithFields("empty image data", logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, whatsapp.ErrMessageSendFailed
	}

	// Get session
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Check if session is connected
	if !sess.IsConnected() {
		uc.logger.WarnWithFields("session not connected", logger.Fields{
			"session_id": sess.ID().String(),
			"status":     sess.Status().String(),
		})
		return nil, session.ErrSessionNotConnected
	}

	// Get WhatsApp client
	waClient, err := uc.waManager.GetClient(sess.ID())
	if err != nil {
		uc.logger.ErrorWithError("failed to get WhatsApp client", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		return nil, whatsapp.ErrClientNotFound
	}

	// Check if client is authenticated
	if !waClient.IsAuthenticated() {
		uc.logger.WarnWithFields("WhatsApp client not authenticated", logger.Fields{
			"session_id": sess.ID().String(),
		})
		return nil, whatsapp.ErrAuthenticationFailed
	}

	// Process image data
	imageData, err := uc.processBase64Image(req.Image)
	if err != nil {
		uc.logger.ErrorWithError("failed to process image data", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
		})
		return &SendImageMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Determine MIME type if not provided
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = uc.detectImageMimeType(imageData)
	}

	// Validate MIME type
	if !uc.isValidImageMimeType(mimeType) {
		err := fmt.Errorf("unsupported image type: %s", mimeType)
		uc.logger.ErrorWithError("invalid image MIME type", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"mime_type":  mimeType,
		})
		return &SendImageMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Validate image size
	const maxImageSize = 16 * 1024 * 1024 // 16MB
	if len(imageData) > maxImageSize {
		err := fmt.Errorf("image too large: %d bytes (max: %d bytes)", len(imageData), maxImageSize)
		uc.logger.ErrorWithError("image size validation failed", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"size":       len(imageData),
		})
		return &SendImageMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Format recipient number
	formattedTo := formatWhatsAppJID(req.To)

	// Send image message (using existing method for now)
	// TODO: Implement SendImageBase64 in whatsapp client
	err = fmt.Errorf("SendImageBase64 not implemented yet")
	if err != nil {
		uc.logger.ErrorWithError("failed to send WhatsApp image message", err, logger.Fields{
			"session_id":  sess.ID().String(),
			"to":          formattedTo,
			"mime_type":   mimeType,
			"has_caption": req.Caption != "",
		})
		return &SendImageMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	uc.logger.InfoWithFields("WhatsApp image message sent successfully", logger.Fields{
		"session_id":  sess.ID().String(),
		"to":          formattedTo,
		"mime_type":   mimeType,
		"image_size":  len(imageData),
		"has_caption": req.Caption != "",
		"has_context": req.ContextInfo != nil,
	})

	return &SendImageMessageResponse{
		SessionID: sess.ID(),
		To:        req.To,
		Success:   true,
		MessageID: generateMessageID(),
	}, nil
}

// processBase64Image processes base64 image data
func (uc *SendImageMessageUseCase) processBase64Image(imageStr string) ([]byte, error) {
	// Remove data URL prefix if present
	if strings.HasPrefix(imageStr, "data:") {
		parts := strings.Split(imageStr, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		imageStr = parts[1]
	}

	// Decode base64
	imageData, err := base64.StdEncoding.DecodeString(imageStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data: %w", err)
	}

	return imageData, nil
}

// detectImageMimeType detects MIME type based on magic bytes
func (uc *SendImageMessageUseCase) detectImageMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// Simple MIME type detection based on magic bytes
	switch {
	case data[0] == 0xFF && data[1] == 0xD8:
		return "image/jpeg"
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
		return "image/png"
	case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
		return "image/gif"
	case data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46:
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// isValidImageMimeType validates if the MIME type is supported for images
func (uc *SendImageMessageUseCase) isValidImageMimeType(mimeType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}

// formatWhatsAppJID formats a phone number to WhatsApp JID format
func formatWhatsAppJID(phone string) string {
	// Remove any non-numeric characters except +
	formatted := strings.ReplaceAll(phone, " ", "")
	formatted = strings.ReplaceAll(formatted, "-", "")
	formatted = strings.ReplaceAll(formatted, "(", "")
	formatted = strings.ReplaceAll(formatted, ")", "")
	formatted = strings.ReplaceAll(formatted, ".", "")

	// Add @s.whatsapp.net if not present
	if !strings.Contains(formatted, "@") {
		// Remove + if present at the beginning
		formatted = strings.TrimPrefix(formatted, "+")
		formatted = formatted + "@s.whatsapp.net"
	}

	return formatted
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	// Generate a random 8-byte ID
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return "msg_" + strings.ReplaceAll(time.Now().Format("20060102_150405"), "_", "")
	}

	return "msg_" + hex.EncodeToString(bytes)
}
