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

// SendAudioMessageUseCase handles sending WhatsApp audio messages
type SendAudioMessageUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
	validator   validator.Validator
}

// NewSendAudioMessageUseCase creates a new send audio message use case
func NewSendAudioMessageUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger, validator validator.Validator) *SendAudioMessageUseCase {
	return &SendAudioMessageUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
		validator:   validator,
	}
}

// SendAudioMessageRequest represents the request to send an audio message
type SendAudioMessageRequest struct {
	SessionID   session.SessionID `json:"session_id"`
	To          string            `json:"to" validate:"required"`
	Audio       string            `json:"audio" validate:"required"` // Base64 string
	MimeType    string            `json:"mime_type"`
	IsPTT       bool              `json:"is_ptt"` // Push-to-talk
	ContextInfo interface{}       `json:"context_info,omitempty"`
}

// SendAudioMessageResponse represents the response from sending an audio message
type SendAudioMessageResponse struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to"`
	Success   bool              `json:"success"`
	MessageID string            `json:"message_id,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// Execute sends an audio message via WhatsApp
func (uc *SendAudioMessageUseCase) Execute(ctx context.Context, req SendAudioMessageRequest) (*SendAudioMessageResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for send audio message", err, logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, err
	}

	// Additional validation for audio data
	if strings.TrimSpace(req.Audio) == "" {
		uc.logger.WarnWithFields("empty audio data", logger.Fields{
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

	// Process audio data
	audioData, err := uc.processBase64Audio(req.Audio)
	if err != nil {
		uc.logger.ErrorWithError("failed to process audio data", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
		})
		return &SendAudioMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Determine MIME type if not provided
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = uc.detectAudioMimeType(audioData)
	}

	// Validate MIME type
	if !uc.isValidAudioMimeType(mimeType) {
		err := fmt.Errorf("unsupported audio type: %s", mimeType)
		uc.logger.ErrorWithError("invalid audio MIME type", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"mime_type":  mimeType,
		})
		return &SendAudioMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Validate audio size
	const maxAudioSize = 16 * 1024 * 1024 // 16MB
	if len(audioData) > maxAudioSize {
		err := fmt.Errorf("audio too large: %d bytes (max: %d bytes)", len(audioData), maxAudioSize)
		uc.logger.ErrorWithError("audio size validation failed", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"size":       len(audioData),
		})
		return &SendAudioMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Format recipient number
	formattedTo := utils.FormatWhatsAppJID(req.To)

	// Send audio message (using placeholder for now)
	// TODO: Implement SendAudioBase64 in whatsapp client
	err = fmt.Errorf("SendAudioBase64 not implemented yet")
	if err != nil {
		uc.logger.ErrorWithError("failed to send WhatsApp audio message", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         formattedTo,
			"mime_type":  mimeType,
			"is_ptt":     req.IsPTT,
		})
		return &SendAudioMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	uc.logger.InfoWithFields("WhatsApp audio message sent successfully", logger.Fields{
		"session_id":  sess.ID().String(),
		"to":          formattedTo,
		"mime_type":   mimeType,
		"audio_size":  len(audioData),
		"is_ptt":      req.IsPTT,
		"has_context": req.ContextInfo != nil,
	})

	return &SendAudioMessageResponse{
		SessionID: sess.ID(),
		To:        req.To,
		Success:   true,
		MessageID: utils.GenerateMessageID(),
	}, nil
}

// processBase64Audio processes base64 audio data
func (uc *SendAudioMessageUseCase) processBase64Audio(audioStr string) ([]byte, error) {
	// Remove data URL prefix if present
	if strings.HasPrefix(audioStr, "data:") {
		parts := strings.Split(audioStr, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		audioStr = parts[1]
	}

	// Decode base64
	audioData, err := base64.StdEncoding.DecodeString(audioStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data: %w", err)
	}

	return audioData, nil
}

// detectAudioMimeType detects MIME type based on magic bytes
func (uc *SendAudioMessageUseCase) detectAudioMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// Simple MIME type detection based on magic bytes
	switch {
	case data[0] == 0xFF && (data[1]&0xE0) == 0xE0: // MP3
		return "audio/mpeg"
	case data[0] == 0x4F && data[1] == 0x67 && data[2] == 0x67 && data[3] == 0x53: // OGG
		return "audio/ogg"
	case data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46: // WAV
		return "audio/wav"
	case data[0] == 0x66 && data[1] == 0x4C && data[2] == 0x61 && data[3] == 0x43: // FLAC
		return "audio/flac"
	default:
		return "audio/mpeg" // Default to MP3
	}
}

// isValidAudioMimeType validates if the MIME type is supported for audio
func (uc *SendAudioMessageUseCase) isValidAudioMimeType(mimeType string) bool {
	validTypes := []string{
		"audio/mpeg",
		"audio/mp3",
		"audio/ogg",
		"audio/wav",
		"audio/flac",
		"audio/aac",
		"audio/m4a",
	}

	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}
