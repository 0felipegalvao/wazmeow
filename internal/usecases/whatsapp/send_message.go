package whatsapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// SendMessageUseCase handles sending WhatsApp messages
type SendMessageUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
	validator   validator.Validator
}

// NewSendMessageUseCase creates a new send message use case
func NewSendMessageUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger, validator validator.Validator) *SendMessageUseCase {
	return &SendMessageUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
		validator:   validator,
	}
}

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to" validate:"required"`
	Message   string            `json:"message" validate:"required,max=4096"`
}

// SendMessageResponse represents the response from sending a message
type SendMessageResponse struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to"`
	Message   string            `json:"message"`
	Success   bool              `json:"success"`
	MessageID string            `json:"message_id,omitempty"`
}

// Execute sends a WhatsApp message
func (uc *SendMessageUseCase) Execute(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for send message", err, logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, err
	}

	// Additional validation
	if strings.TrimSpace(req.Message) == "" {
		uc.logger.WarnWithFields("empty message content", logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, whatsapp.ErrMessageSendFailed
	}

	// Get session from repository
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
		uc.logger.ErrorWithError("WhatsApp client not found", err, logger.Fields{
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

	// Format recipient number
	formattedTo := formatRecipient(req.To)

	// Send message
	err = waClient.SendMessage(ctx, formattedTo, req.Message)
	if err != nil {
		uc.logger.ErrorWithError("failed to send WhatsApp message", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         formattedTo,
			"message":    truncateMessage(req.Message, 100),
		})
		return &SendMessageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			Message:   req.Message,
			Success:   false,
		}, err
	}

	uc.logger.InfoWithFields("WhatsApp message sent successfully", logger.Fields{
		"session_id":     sess.ID().String(),
		"to":             formattedTo,
		"message_length": len(req.Message),
	})

	return &SendMessageResponse{
		SessionID: sess.ID(),
		To:        req.To,
		Message:   req.Message,
		Success:   true,
		MessageID: generateMessageID(), // In real implementation, this would come from WhatsApp
	}, nil
}

// SendBulkMessageRequest represents the request to send messages to multiple recipients
type SendBulkMessageRequest struct {
	SessionID session.SessionID `json:"session_id"`
	To        []string          `json:"to" validate:"required,min=1,max=100"`
	Message   string            `json:"message" validate:"required,max=4096"`
}

// SendBulkMessageResponse represents the response from sending bulk messages
type SendBulkMessageResponse struct {
	SessionID    session.SessionID     `json:"session_id"`
	Message      string                `json:"message"`
	TotalCount   int                   `json:"total_count"`
	SuccessCount int                   `json:"success_count"`
	FailedCount  int                   `json:"failed_count"`
	Results      []SendMessageResponse `json:"results"`
	Errors       []string              `json:"errors,omitempty"`
}

// ExecuteBulk sends a message to multiple recipients
func (uc *SendMessageUseCase) ExecuteBulk(ctx context.Context, req SendBulkMessageRequest) (*SendBulkMessageResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for bulk send message", err, logger.Fields{
			"session_id":      req.SessionID.String(),
			"recipient_count": len(req.To),
		})
		return nil, err
	}

	response := &SendBulkMessageResponse{
		SessionID:  req.SessionID,
		Message:    req.Message,
		TotalCount: len(req.To),
		Results:    make([]SendMessageResponse, 0, len(req.To)),
	}

	var errors []string

	// Send message to each recipient
	for _, recipient := range req.To {
		sendReq := SendMessageRequest{
			SessionID: req.SessionID,
			To:        recipient,
			Message:   req.Message,
		}

		result, err := uc.Execute(ctx, sendReq)
		if err != nil {
			response.FailedCount++
			errorMsg := fmt.Sprintf("Failed to send to %s: %v", recipient, err)
			errors = append(errors, errorMsg)

			// Add failed result
			response.Results = append(response.Results, SendMessageResponse{
				SessionID: req.SessionID,
				To:        recipient,
				Message:   req.Message,
				Success:   false,
			})
		} else {
			response.SuccessCount++
			response.Results = append(response.Results, *result)
		}
	}

	response.Errors = errors

	uc.logger.InfoWithFields("bulk message sending completed", logger.Fields{
		"session_id":    req.SessionID.String(),
		"total_count":   response.TotalCount,
		"success_count": response.SuccessCount,
		"failed_count":  response.FailedCount,
	})

	return response, nil
}

// SendImageRequest represents the request to send an image
type SendImageRequest struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to" validate:"required"`
	ImagePath string            `json:"image_path" validate:"required"`
	Caption   string            `json:"caption,omitempty" validate:"max=1024"`
}

// SendImageResponse represents the response from sending an image
type SendImageResponse struct {
	SessionID session.SessionID `json:"session_id"`
	To        string            `json:"to"`
	ImagePath string            `json:"image_path"`
	Caption   string            `json:"caption,omitempty"`
	Success   bool              `json:"success"`
	MessageID string            `json:"message_id,omitempty"`
}

// ExecuteSendImage sends an image message
func (uc *SendMessageUseCase) ExecuteSendImage(ctx context.Context, req SendImageRequest) (*SendImageResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for send image", err, logger.Fields{
			"session_id": req.SessionID.String(),
			"to":         req.To,
		})
		return nil, err
	}

	// Get session and validate connection (similar to text message)
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if !sess.IsConnected() {
		return nil, session.ErrSessionNotConnected
	}

	waClient, err := uc.waManager.GetClient(sess.ID())
	if err != nil {
		return nil, whatsapp.ErrClientNotFound
	}

	if !waClient.IsAuthenticated() {
		return nil, whatsapp.ErrAuthenticationFailed
	}

	// Format recipient number
	formattedTo := formatRecipient(req.To)

	// Send image
	err = waClient.SendImage(ctx, formattedTo, req.ImagePath, req.Caption)
	if err != nil {
		uc.logger.ErrorWithError("failed to send WhatsApp image", err, logger.Fields{
			"session_id": sess.ID().String(),
			"to":         formattedTo,
			"image_path": req.ImagePath,
		})
		return &SendImageResponse{
			SessionID: sess.ID(),
			To:        req.To,
			ImagePath: req.ImagePath,
			Caption:   req.Caption,
			Success:   false,
		}, err
	}

	uc.logger.InfoWithFields("WhatsApp image sent successfully", logger.Fields{
		"session_id": sess.ID().String(),
		"to":         formattedTo,
		"image_path": req.ImagePath,
	})

	return &SendImageResponse{
		SessionID: sess.ID(),
		To:        req.To,
		ImagePath: req.ImagePath,
		Caption:   req.Caption,
		Success:   true,
		MessageID: generateMessageID(),
	}, nil
}

// Helper functions

// formatRecipient formats a recipient number to WhatsApp JID format
func formatRecipient(recipient string) string {
	// Remove any non-digit characters except +
	cleaned := strings.ReplaceAll(recipient, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// If it doesn't contain @, assume it's a phone number and add @s.whatsapp.net
	if !strings.Contains(cleaned, "@") {
		// Remove + if present
		cleaned = strings.TrimPrefix(cleaned, "+")
		return cleaned + "@s.whatsapp.net"
	}

	return cleaned
}

// truncateMessage truncates a message for logging purposes
func truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength] + "..."
}

// generateMessageID generates a mock message ID
func generateMessageID() string {
	// In real implementation, this would come from WhatsApp
	return "mock-message-id-" + fmt.Sprintf("%d", time.Now().Unix())
}
