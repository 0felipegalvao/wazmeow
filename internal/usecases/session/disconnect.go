package session

import (
	"context"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
)

// DisconnectUseCase handles session disconnection from WhatsApp
type DisconnectUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
}

// NewDisconnectUseCase creates a new disconnect session use case
func NewDisconnectUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger) *DisconnectUseCase {
	return &DisconnectUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
	}
}

// DisconnectRequest represents the request to disconnect a session
type DisconnectRequest struct {
	SessionID session.SessionID `json:"session_id"`
}

// DisconnectResponse represents the response from disconnecting a session
type DisconnectResponse struct {
	Session *session.Session `json:"session"`
	Message string           `json:"message"`
}

// Execute disconnects a session from WhatsApp
func (uc *DisconnectUseCase) Execute(ctx context.Context, req DisconnectRequest) (*DisconnectResponse, error) {
	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Check if session is connected
	if sess.Status() == session.StatusDisconnected {
		uc.logger.InfoWithFields("session already disconnected", logger.Fields{
			"session_id": sess.ID().String(),
		})
		return &DisconnectResponse{
			Session: sess,
			Message: "Session already disconnected",
		}, nil
	}

	// Get WhatsApp client if it exists
	waClient, err := uc.waManager.GetClient(sess.ID())
	if err != nil {
		// Client doesn't exist, just update session status
		uc.logger.WarnWithFields("WhatsApp client not found, updating session status only", logger.Fields{
			"session_id": sess.ID().String(),
		})
	} else {
		// Disconnect from WhatsApp
		if err := waClient.Disconnect(ctx); err != nil {
			uc.logger.ErrorWithError("failed to disconnect from WhatsApp", err, logger.Fields{
				"session_id": sess.ID().String(),
			})
			// Continue with session update even if WhatsApp disconnect fails
		}
	}

	// Update session status
	sess.Disconnect()
	if err := uc.sessionRepo.Update(ctx, sess); err != nil {
		uc.logger.ErrorWithError("failed to update session status", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		return nil, err
	}

	uc.logger.InfoWithFields("session disconnected successfully", logger.Fields{
		"session_id": sess.ID().String(),
		"status":     sess.Status().String(),
	})

	return &DisconnectResponse{
		Session: sess,
		Message: "Session disconnected successfully",
	}, nil
}
