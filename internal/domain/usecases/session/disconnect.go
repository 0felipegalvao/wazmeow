package usecase

import (
	"context"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"

	"github.com/rs/zerolog/log"
)

// DisconnectSessionRequest represents the request to disconnect a session
type DisconnectSessionRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	Logout    bool   `json:"logout,omitempty"` // If true, also logout and clear authentication
}

// DisconnectSessionResponse represents the response after disconnecting a session
type DisconnectSessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// DisconnectSessionUseCase handles disconnecting sessions from WhatsApp
type DisconnectSessionUseCase struct {
	sessionRepo    session.Repository
	whatsappClient whatsapp.ClientManager
}

// NewDisconnectSessionUseCase creates a new instance of DisconnectSessionUseCase
func NewDisconnectSessionUseCase(sessionRepo session.Repository, whatsappClient whatsapp.ClientManager) *DisconnectSessionUseCase {
	return &DisconnectSessionUseCase{
		sessionRepo:    sessionRepo,
		whatsappClient: whatsappClient,
	}
}

// Execute disconnects a session from WhatsApp
func (uc *DisconnectSessionUseCase) Execute(ctx context.Context, req DisconnectSessionRequest) (*DisconnectSessionResponse, error) {
	// Validate session ID
	sessionID := session.SessionID(req.SessionID)
	if !sessionID.IsValid() {
		return nil, session.ErrInvalidSessionName("invalid session ID format")
	}

	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to get session")
		return nil, err
	}

	// Check if session can be disconnected
	if !sess.CanDisconnect() {
		return nil, session.ErrCannotDisconnect(sessionID, sess.Status())
	}

	// Get WhatsApp client
	client, err := uc.whatsappClient.GetClient(ctx, sessionID)
	if err != nil {
		log.Warn().Err(err).Str("session_id", req.SessionID).Msg("WhatsApp client not found, marking session as disconnected")

		// If client doesn't exist, just mark session as disconnected
		sess.Disconnect()
		if err := uc.sessionRepo.Update(ctx, sess); err != nil {
			log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to update session status")
			return nil, err
		}

		return &DisconnectSessionResponse{
			SessionID: req.SessionID,
			Status:    string(sess.Status()),
			Message:   "Session marked as disconnected",
		}, nil
	}

	// Perform logout if requested
	if req.Logout {
		if err := client.Logout(ctx, sessionID); err != nil {
			log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to logout from WhatsApp")
			// Continue with disconnect even if logout fails
		} else {
			// Clear authentication data
			sess.ClearAuthentication()
		}
	} else {
		// Just disconnect
		if err := client.Disconnect(ctx, sessionID); err != nil {
			log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to disconnect from WhatsApp")
			// Continue to mark as disconnected even if disconnect fails
		}
	}

	// Update session status
	sess.Disconnect()
	if err := uc.sessionRepo.Update(ctx, sess); err != nil {
		log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to update session status")
		return nil, err
	}

	// Remove client if logout was performed
	if req.Logout {
		if err := uc.whatsappClient.RemoveClient(ctx, sessionID); err != nil {
			log.Warn().Err(err).Str("session_id", req.SessionID).Msg("Failed to remove WhatsApp client")
		}
	}

	message := "Session disconnected successfully"
	if req.Logout {
		message = "Session logged out and disconnected successfully"
	}

	log.Info().
		Str("session_id", req.SessionID).
		Bool("logout", req.Logout).
		Str("status", string(sess.Status())).
		Msg("Session disconnected")

	return &DisconnectSessionResponse{
		SessionID: req.SessionID,
		Status:    string(sess.Status()),
		Message:   message,
	}, nil
}
