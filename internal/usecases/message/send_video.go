package message

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/internal/shared/utils"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// SendVideoMessageUseCase handles sending WhatsApp video messages
type SendVideoMessageUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
	validator   validator.Validator
}

// NewSendVideoMessageUseCase creates a new send video message use case
func NewSendVideoMessageUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger, validator validator.Validator) *SendVideoMessageUseCase {
	return &SendVideoMessageUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
		validator:   validator,
	}
}

// SendVideoMessageRequest represents the request to send a video message
type SendVideoMessageRequest struct {
	SessionID   session.SessionID `json:"session_id"`
	To          string            `json:"to" validate:"required"`
	Video       string            `json:"video" validate:"required"` // Base64 string
	Caption     string            `json:"caption" validate:"max=1024"`
	MimeType    string            `json:"mime_type"`
	ContextInfo interface{}       `json:"context_info,omitempty"`
}

// SendVideoMessageResponse represents the response from sending a video message
type SendVideoMessageResponse struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to"`
	Success   bool              `json:"success"`
	MessageID string            `json:"message_id,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// Execute sends a video message via WhatsApp
func (uc *SendVideoMessageUseCase) Execute(ctx context.Context, req SendVideoMessageRequest) (*SendVideoMessageResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for send video message", err, logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, err
	}

	// Additional validation for video data
	if strings.TrimSpace(req.Video) == "" {
		uc.logger.WarnWithFields("empty video data", logger.Fields{
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

	// Process video data
	videoData, err := uc.processBase64Video(req.Video)
	if err != nil {
		uc.logger.ErrorWithError("failed to process video data", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
		})
		return &SendVideoMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Determine MIME type if not provided
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = uc.detectVideoMimeType(videoData)
	}

	// Validate MIME type
	if !uc.isValidVideoMimeType(mimeType) {
		err := fmt.Errorf("unsupported video type: %s", mimeType)
		uc.logger.ErrorWithError("invalid video MIME type", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"mime_type":  mimeType,
		})
		return &SendVideoMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Validate video size
	const maxVideoSize = 64 * 1024 * 1024 // 64MB
	if len(videoData) > maxVideoSize {
		err := fmt.Errorf("video too large: %d bytes (max: %d bytes)", len(videoData), maxVideoSize)
		uc.logger.ErrorWithError("video size validation failed", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"size":       len(videoData),
		})
		return &SendVideoMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Format recipient number
	formattedTo := utils.FormatWhatsAppJID(req.To)

	// Send video message (using placeholder for now)
	// TODO: Implement SendVideoBase64 in whatsapp client
	err = fmt.Errorf("SendVideoBase64 not implemented yet")
	if err != nil {
		uc.logger.ErrorWithError("failed to send WhatsApp video message", err, logger.Fields{
			"session_id":  sess.ID().String(),
			"to":          formattedTo,
			"mime_type":   mimeType,
			"has_caption": req.Caption != "",
		})
		return &SendVideoMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	uc.logger.InfoWithFields("WhatsApp video message sent successfully", logger.Fields{
		"session_id":  sess.ID().String(),
		"to":          formattedTo,
		"mime_type":   mimeType,
		"video_size":  len(videoData),
		"has_caption": req.Caption != "",
		"has_context": req.ContextInfo != nil,
	})

	return &SendVideoMessageResponse{
		SessionID: sess.ID(),
		To:        req.To,
		Success:   true,
		MessageID: utils.GenerateMessageID(),
	}, nil
}

// processBase64Video processes base64 video data
func (uc *SendVideoMessageUseCase) processBase64Video(videoStr string) ([]byte, error) {
	// Remove data URL prefix if present
	if strings.HasPrefix(videoStr, "data:") {
		parts := strings.Split(videoStr, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		videoStr = parts[1]
	}

	// Decode base64
	videoData, err := base64.StdEncoding.DecodeString(videoStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data: %w", err)
	}

	return videoData, nil
}

// detectVideoMimeType detects MIME type based on magic bytes
func (uc *SendVideoMessageUseCase) detectVideoMimeType(data []byte) string {
	if len(data) < 8 {
		return "application/octet-stream"
	}

	// Simple MIME type detection based on magic bytes
	switch {
	case data[4] == 0x66 && data[5] == 0x74 && data[6] == 0x79 && data[7] == 0x70: // MP4
		return "video/mp4"
	case data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3: // WebM/MKV
		return "video/webm"
	case data[0] == 0x46 && data[1] == 0x4C && data[2] == 0x56: // FLV
		return "video/x-flv"
	case data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 && data[3] == 0xBA: // MPEG
		return "video/mpeg"
	default:
		return "video/mp4" // Default to MP4
	}
}

// isValidVideoMimeType validates if the MIME type is supported for videos
func (uc *SendVideoMessageUseCase) isValidVideoMimeType(mimeType string) bool {
	validTypes := []string{
		"video/mp4",
		"video/mpeg",
		"video/webm",
		"video/quicktime",
		"video/x-msvideo", // AVI
		"video/x-flv",
	}

	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}
