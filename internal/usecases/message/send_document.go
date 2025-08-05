package message

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/internal/shared/utils"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// SendDocumentMessageUseCase handles sending WhatsApp document messages
type SendDocumentMessageUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
	validator   validator.Validator
}

// NewSendDocumentMessageUseCase creates a new send document message use case
func NewSendDocumentMessageUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger, validator validator.Validator) *SendDocumentMessageUseCase {
	return &SendDocumentMessageUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
		validator:   validator,
	}
}

// SendDocumentMessageRequest represents the request to send a document message
type SendDocumentMessageRequest struct {
	SessionID   session.SessionID `json:"session_id"`
	To          string            `json:"to" validate:"required"`
	Document    string            `json:"document" validate:"required"` // Base64 string
	Filename    string            `json:"filename" validate:"required"`
	MimeType    string            `json:"mime_type"`
	ContextInfo interface{}       `json:"context_info,omitempty"`
}

// SendDocumentMessageResponse represents the response from sending a document message
type SendDocumentMessageResponse struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to"`
	Filename  string            `json:"filename"`
	Success   bool              `json:"success"`
	MessageID string            `json:"message_id,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// Execute sends a document message via WhatsApp
func (uc *SendDocumentMessageUseCase) Execute(ctx context.Context, req SendDocumentMessageRequest) (*SendDocumentMessageResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for send document message", err, logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
			"filename":   req.Filename,
		})
		return nil, err
	}

	// Additional validation for document data
	if strings.TrimSpace(req.Document) == "" {
		uc.logger.WarnWithFields("empty document data", logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
			"filename":   req.Filename,
		})
		return nil, whatsapp.ErrMessageSendFailed
	}

	// Validate filename
	if strings.TrimSpace(req.Filename) == "" {
		uc.logger.WarnWithFields("empty filename", logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, fmt.Errorf("filename is required")
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

	// Process document data
	documentData, err := uc.processBase64Document(req.Document)
	if err != nil {
		uc.logger.ErrorWithError("failed to process document data", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"filename":   req.Filename,
		})
		return &SendDocumentMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Filename:  req.Filename,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Determine MIME type if not provided
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = uc.detectDocumentMimeType(documentData, req.Filename)
	}

	// Validate MIME type
	if !uc.isValidDocumentMimeType(mimeType) {
		err := fmt.Errorf("unsupported document type: %s", mimeType)
		uc.logger.ErrorWithError("invalid document MIME type", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"filename":   req.Filename,
			"mime_type":  mimeType,
		})
		return &SendDocumentMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Filename:  req.Filename,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Validate document size
	const maxDocumentSize = 100 * 1024 * 1024 // 100MB
	if len(documentData) > maxDocumentSize {
		err := fmt.Errorf("document too large: %d bytes (max: %d bytes)", len(documentData), maxDocumentSize)
		uc.logger.ErrorWithError("document size validation failed", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         req.To,
			"filename":   req.Filename,
			"size":       len(documentData),
		})
		return &SendDocumentMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Filename:  req.Filename,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	// Format recipient number
	formattedTo := utils.FormatWhatsAppJID(req.To)

	// Send document message (using placeholder for now)
	// TODO: Implement SendDocumentBase64 in whatsapp client
	err = fmt.Errorf("SendDocumentBase64 not implemented yet")
	if err != nil {
		uc.logger.ErrorWithError("failed to send WhatsApp document message", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         formattedTo,
			"filename":   req.Filename,
			"mime_type":  mimeType,
		})
		return &SendDocumentMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Filename:  req.Filename,
			Success:   false,
			Error:     err.Error(),
		}, err
	}

	uc.logger.InfoWithFields("WhatsApp document message sent successfully", logger.Fields{
		"session_id":    sess.ID().String(),
		"to":            formattedTo,
		"filename":      req.Filename,
		"mime_type":     mimeType,
		"document_size": len(documentData),
		"has_context":   req.ContextInfo != nil,
	})

	return &SendDocumentMessageResponse{
		SessionID: sess.ID(),
		To:        req.To,
		Filename:  req.Filename,
		Success:   true,
		MessageID: utils.GenerateMessageID(),
	}, nil
}

// processBase64Document processes base64 document data
func (uc *SendDocumentMessageUseCase) processBase64Document(documentStr string) ([]byte, error) {
	// Remove data URL prefix if present
	if strings.HasPrefix(documentStr, "data:") {
		parts := strings.Split(documentStr, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		documentStr = parts[1]
	}

	// Decode base64
	documentData, err := base64.StdEncoding.DecodeString(documentStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data: %w", err)
	}

	return documentData, nil
}

// detectDocumentMimeType detects MIME type based on file extension and magic bytes
func (uc *SendDocumentMessageUseCase) detectDocumentMimeType(data []byte, filename string) string {
	// First try to detect by file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".7z":
		return "application/x-7z-compressed"
	}

	// Try to detect by magic bytes
	if len(data) >= 4 {
		switch {
		case data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46: // PDF
			return "application/pdf"
		case data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04: // ZIP/Office
			return "application/zip"
		case data[0] == 0xD0 && data[1] == 0xCF && data[2] == 0x11 && data[3] == 0xE0: // MS Office
			return "application/msword"
		}
	}

	return "application/octet-stream"
}

// isValidDocumentMimeType validates if the MIME type is supported for documents
func (uc *SendDocumentMessageUseCase) isValidDocumentMimeType(mimeType string) bool {
	// Allow most document types - WhatsApp is quite permissive
	restrictedTypes := []string{
		"application/x-executable",
		"application/x-msdownload",
		"application/x-dosexec",
	}

	for _, restrictedType := range restrictedTypes {
		if mimeType == restrictedType {
			return false
		}
	}

	return true // Allow most document types
}
