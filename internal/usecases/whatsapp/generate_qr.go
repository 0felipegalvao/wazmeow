package whatsapp

import (
	"context"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
)

// GenerateQRUseCase handles QR code generation for WhatsApp authentication
type GenerateQRUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
}

// NewGenerateQRUseCase creates a new generate QR use case
func NewGenerateQRUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger) *GenerateQRUseCase {
	return &GenerateQRUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
	}
}

// GenerateQRRequest represents the request to generate a QR code
type GenerateQRRequest struct {
	SessionID session.SessionID `json:"session_id"`
}

// GenerateQRResponse represents the response from generating a QR code
type GenerateQRResponse struct {
	SessionID session.SessionID `json:"session_id"`
	QRCode    string            `json:"qr_code"`
	Message   string            `json:"message"`
}

// Execute generates a QR code for WhatsApp authentication
func (uc *GenerateQRUseCase) Execute(ctx context.Context, req GenerateQRRequest) (*GenerateQRResponse, error) {
	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Check if session is in a valid state for QR generation
	if sess.IsConnected() {
		uc.logger.WarnWithFields("session already connected", logger.Fields{
			"session_id": sess.ID().String(),
			"status":     sess.Status().String(),
		})
		return nil, session.ErrSessionAlreadyConnected
	}

	// Check if session already has a QR code saved
	if sess.QRCode() != "" {
		uc.logger.InfoWithFields("returning saved QR code from database", logger.Fields{
			"session_id": sess.ID().String(),
			"qr_length":  len(sess.QRCode()),
		})
		return &GenerateQRResponse{
			SessionID: sess.ID(),
			QRCode:    sess.QRCode(),
			Message:   "QR code retrieved from database. Scan with WhatsApp mobile app.",
		}, nil
	}

	// Get WhatsApp client
	waClient, err := uc.waManager.GetClient(sess.ID())
	if err != nil {
		// Create client if it doesn't exist
		waClient, err = uc.waManager.CreateClient(sess.ID())
		if err != nil {
			uc.logger.ErrorWithError("failed to create WhatsApp client", err, logger.Fields{
				"session_id": sess.ID().String(),
			})
			return nil, err
		}
	}

	// Check if already authenticated
	if waClient.IsAuthenticated() {
		uc.logger.InfoWithFields("session already authenticated", logger.Fields{
			"session_id": sess.ID().String(),
			"jid":        waClient.GetJID(),
		})
		return &GenerateQRResponse{
			SessionID: sess.ID(),
			Message:   "Session already authenticated",
		}, nil
	}

	// Generate QR code
	qrCode, err := waClient.GenerateQR(ctx)
	if err != nil {
		uc.logger.ErrorWithError("failed to generate QR code", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		return nil, err
	}

	uc.logger.InfoWithFields("QR code generated successfully", logger.Fields{
		"session_id": sess.ID().String(),
		"qr_length":  len(qrCode),
	})

	return &GenerateQRResponse{
		SessionID: sess.ID(),
		QRCode:    qrCode,
		Message:   "QR code generated successfully. Scan with WhatsApp mobile app.",
	}, nil
}

// RefreshQRRequest represents the request to refresh a QR code
type RefreshQRRequest struct {
	SessionID session.SessionID `json:"session_id"`
}

// RefreshQRResponse represents the response from refreshing a QR code
type RefreshQRResponse struct {
	SessionID session.SessionID `json:"session_id"`
	QRCode    string            `json:"qr_code"`
	Message   string            `json:"message"`
}

// ExecuteRefresh refreshes an existing QR code
func (uc *GenerateQRUseCase) ExecuteRefresh(ctx context.Context, req RefreshQRRequest) (*RefreshQRResponse, error) {
	// Get session from repository
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session for QR refresh", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Get WhatsApp client
	waClient, err := uc.waManager.GetClient(sess.ID())
	if err != nil {
		uc.logger.ErrorWithError("WhatsApp client not found for QR refresh", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		return nil, whatsapp.ErrClientNotFound
	}

	// Check if already authenticated
	if waClient.IsAuthenticated() {
		uc.logger.InfoWithFields("session already authenticated, cannot refresh QR", logger.Fields{
			"session_id": sess.ID().String(),
			"jid":        waClient.GetJID(),
		})
		return &RefreshQRResponse{
			SessionID: sess.ID(),
			Message:   "Session already authenticated",
		}, nil
	}

	// Generate new QR code
	qrCode, err := waClient.GenerateQR(ctx)
	if err != nil {
		uc.logger.ErrorWithError("failed to refresh QR code", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		return nil, err
	}

	uc.logger.InfoWithFields("QR code refreshed successfully", logger.Fields{
		"session_id": sess.ID().String(),
		"qr_length":  len(qrCode),
	})

	return &RefreshQRResponse{
		SessionID: sess.ID(),
		QRCode:    qrCode,
		Message:   "QR code refreshed successfully. Scan with WhatsApp mobile app.",
	}, nil
}
