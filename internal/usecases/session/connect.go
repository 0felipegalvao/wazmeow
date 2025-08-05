package session

import (
	"context"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
)

// ConnectUseCase handles session connection to WhatsApp
type ConnectUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
}

// NewConnectUseCase creates a new connect session use case
func NewConnectUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger) *ConnectUseCase {
	return &ConnectUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
	}
}

// ConnectRequest represents the request to connect a session
type ConnectRequest struct {
	SessionID session.SessionID `json:"session_id"`
}

// ConnectResponse represents the response from connecting a session
type ConnectResponse struct {
	Session   *session.Session `json:"session"`
	QRCode    string           `json:"qr_code,omitempty"`
	NeedsAuth bool             `json:"needs_auth"`
	Message   string           `json:"message"`
}

// Execute connects a session to WhatsApp
func (uc *ConnectUseCase) Execute(ctx context.Context, req ConnectRequest) (*ConnectResponse, error) {
	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Check if session is already connected
	if sess.Status() == session.StatusConnected {
		uc.logger.WarnWithFields("session already connected", logger.Fields{
			"session_id": sess.ID().String(),
			"status":     sess.Status().String(),
		})
		return nil, session.ErrSessionAlreadyConnected
	}

	// Log if reconnecting from connecting state
	if sess.Status() == session.StatusConnecting {
		uc.logger.InfoWithFields("reconnecting session that was in connecting state", logger.Fields{
			"session_id": sess.ID().String(),
			"status":     sess.Status().String(),
		})
	}

	// Check if session can be connected (allows disconnecting, connecting, disconnected)
	if !sess.CanConnect() {
		uc.logger.WarnWithFields("session cannot be connected", logger.Fields{
			"session_id": sess.ID().String(),
			"status":     sess.Status().String(),
		})
		return nil, session.ErrSessionInvalidState
	}

	// Set session to connecting state
	sess.SetConnecting()
	if err := uc.sessionRepo.Update(ctx, sess); err != nil {
		uc.logger.ErrorWithError("failed to update session status", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		return nil, err
	}

	// Get or create WhatsApp client
	waClient, err := uc.waManager.GetClient(sess.ID())
	if err != nil {
		// Create new client if it doesn't exist
		waClient, err = uc.waManager.CreateClient(sess.ID())
		if err != nil {
			uc.logger.ErrorWithError("failed to create WhatsApp client", err, logger.Fields{
				"session_id": sess.ID().String(),
			})
			sess.Disconnect()
			uc.sessionRepo.Update(ctx, sess)
			return nil, err
		}
	}

	// Attempt to connect
	result, err := waClient.Connect(ctx)
	if err != nil {
		uc.logger.ErrorWithError("failed to connect to WhatsApp", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		sess.Disconnect()
		uc.sessionRepo.Update(ctx, sess)
		return nil, err
	}

	response := &ConnectResponse{
		Session: sess,
	}

	// Handle connection result and map to session status
	switch result.Status {
	case whatsapp.StatusConnected, whatsapp.StatusAuthenticated:
		if result.JID != "" {
			// Already authenticated - mark session as connected
			if err := sess.Connect(result.JID); err != nil {
				uc.logger.ErrorWithError("failed to update session with JID", err, logger.Fields{
					"session_id": sess.ID().String(),
					"jid":        result.JID,
				})
				return nil, err
			}
			if err := uc.sessionRepo.Update(ctx, sess); err != nil {
				return nil, err
			}
			response.Message = "Connected and authenticated successfully"
		} else {
			// Connected but not authenticated yet - mark as connecting
			sess.SetConnecting()
			if err := uc.sessionRepo.Update(ctx, sess); err != nil {
				return nil, err
			}
			response.Message = "Connected, waiting for authentication"
		}

	case whatsapp.StatusAuthenticating:
		// Need authentication (QR code or pairing) - mark as connecting
		sess.SetConnecting()
		if err := uc.sessionRepo.Update(ctx, sess); err != nil {
			return nil, err
		}
		response.NeedsAuth = true
		response.QRCode = result.QRCode
		response.Message = "Connection established, authentication required"

	case whatsapp.StatusConnecting:
		// Still connecting - mark as connecting
		sess.SetConnecting()
		if err := uc.sessionRepo.Update(ctx, sess); err != nil {
			return nil, err
		}
		response.Message = "Connection in progress"

	default:
		// Any other status (including error) - mark as disconnected
		sess.Disconnect()
		if err := uc.sessionRepo.Update(ctx, sess); err != nil {
			return nil, err
		}
		response.Message = "Connection failed"
	}

	uc.logger.InfoWithFields("session connection processed", logger.Fields{
		"session_id": sess.ID().String(),
		"status":     result.Status.String(),
		"needs_auth": response.NeedsAuth,
		"has_qr":     response.QRCode != "",
	})

	return response, nil
}
