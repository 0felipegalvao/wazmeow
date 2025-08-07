package usecase

import (
	"context"

	"github.com/rs/zerolog/log"
)

// ConnectSessionRequest represents the request to connect a session
type ConnectSessionRequest struct {
	SessionID string   `json:"session_id" validate:"required"`
	Events    []string `json:"events,omitempty"`
	Immediate bool     `json:"immediate,omitempty"`
}

// ConnectSessionResponse represents the response after connecting a session
type ConnectSessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// ConnectSessionUseCase handles connecting sessions to WhatsApp
type ConnectSessionUseCase struct {
	sessionRepo    Repository
	whatsappClient ClientManager
}

// NewConnectSessionUseCase creates a new instance of ConnectSessionUseCase
func NewConnectSessionUseCase(sessionRepo Repository, whatsappClient ClientManager) *ConnectSessionUseCase {
	return &ConnectSessionUseCase{
		sessionRepo:    sessionRepo,
		whatsappClient: whatsappClient,
	}
}

// Execute connects a session to WhatsApp
func (uc *ConnectSessionUseCase) Execute(ctx context.Context, req ConnectSessionRequest) (*ConnectSessionResponse, error) {
	// Validate session ID
	sessionID := SessionID(req.SessionID)
	if !sessionID.IsValid() {
		return nil, ErrInvalidSessionName("invalid session ID format")
	}

	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to get session")
		return nil, err
	}

	// Check if session can be connected
	if !sess.CanConnect() {
		return nil, ErrCannotConnect(sessionID, sess.Status())
	}

	// Update session status to connecting
	if err := sess.Connect(); err != nil {
		return nil, err
	}

	// Save status update
	if err := uc.sessionRepo.Update(ctx, sess); err != nil {
		log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to update session status")
		return nil, err
	}

	// Create or get WhatsApp client
	_, err = uc.whatsappClient.GetClient(ctx, sessionID)
	if err != nil {
		// Try to create a new client if it doesn't exist
		_, err = uc.whatsappClient.CreateClient(ctx, sessionID)
		if err != nil {
			log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to create WhatsApp client")

			// Mark session as error
			sess.MarkError()
			uc.sessionRepo.Update(ctx, sess)

			return nil, err
		}
	}

	// Connect to WhatsApp using the ClientManager interface
	if err := uc.whatsappClient.ConnectAll(ctx); err != nil {
		log.Error().Err(err).Str("session_id", req.SessionID).Msg("Failed to connect to WhatsApp")

		// Mark session as error
		sess.MarkError()
		uc.sessionRepo.Update(ctx, sess)

		return nil, err
	}

	log.Info().
		Str("session_id", req.SessionID).
		Str("status", string(sess.Status())).
		Msg("Session connection initiated")

	return &ConnectSessionResponse{
		SessionID: req.SessionID,
		Status:    string(sess.Status()),
		Message:   "Connection initiated successfully",
	}, nil
}
