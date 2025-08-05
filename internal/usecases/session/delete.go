package session

import (
	"context"
	"fmt"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
)

// DeleteUseCase handles session deletion
type DeleteUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
}

// NewDeleteUseCase creates a new delete session use case
func NewDeleteUseCase(sessionRepo session.Repository, waManager whatsapp.Manager, logger logger.Logger) *DeleteUseCase {
	return &DeleteUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
	}
}

// DeleteRequest represents the request to delete a session
type DeleteRequest struct {
	SessionID session.SessionID `json:"session_id"`
	Force     bool              `json:"force"` // Force delete even if connected
}

// DeleteResponse represents the response from deleting a session
type DeleteResponse struct {
	SessionID session.SessionID `json:"session_id"`
	Message   string            `json:"message"`
}

// Execute deletes a session
func (uc *DeleteUseCase) Execute(ctx context.Context, req DeleteRequest) (*DeleteResponse, error) {
	// Get session from repository to verify it exists
	sess, err := uc.sessionRepo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session for deletion", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Check if session is connected and force is not set
	if sess.IsConnected() && !req.Force {
		uc.logger.WarnWithFields("cannot delete connected session without force flag", logger.Fields{
			"session_id": sess.ID().String(),
			"status":     sess.Status().String(),
		})
		return nil, session.ErrSessionInvalidState
	}

	// If session is connected, disconnect it first
	if sess.IsConnected() {
		uc.logger.InfoWithFields("disconnecting session before deletion", logger.Fields{
			"session_id": sess.ID().String(),
		})

		// Get WhatsApp client and disconnect
		waClient, err := uc.waManager.GetClient(sess.ID())
		if err == nil {
			if err := waClient.Disconnect(ctx); err != nil {
				uc.logger.ErrorWithError("failed to disconnect WhatsApp client", err, logger.Fields{
					"session_id": sess.ID().String(),
				})
				// Continue with deletion even if disconnect fails
			}
		}

		// Remove WhatsApp client
		if err := uc.waManager.RemoveClient(sess.ID()); err != nil {
			uc.logger.ErrorWithError("failed to remove WhatsApp client", err, logger.Fields{
				"session_id": sess.ID().String(),
			})
			// Continue with deletion even if client removal fails
		}
	}

	// Delete session from repository
	if err := uc.sessionRepo.Delete(ctx, req.SessionID); err != nil {
		uc.logger.ErrorWithError("failed to delete session from repository", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	uc.logger.InfoWithFields("session deleted successfully", logger.Fields{
		"session_id": req.SessionID.String(),
		"name":       sess.Name(),
		"force":      req.Force,
	})

	return &DeleteResponse{
		SessionID: req.SessionID,
		Message:   "Session deleted successfully",
	}, nil
}

// DeleteAllRequest represents the request to delete all sessions
type DeleteAllRequest struct {
	Force bool `json:"force"` // Force delete even if some are connected
}

// DeleteAllResponse represents the response from deleting all sessions
type DeleteAllResponse struct {
	DeletedCount int      `json:"deleted_count"`
	FailedCount  int      `json:"failed_count"`
	Errors       []string `json:"errors,omitempty"`
	Message      string   `json:"message"`
}

// ExecuteDeleteAll deletes all sessions
func (uc *DeleteUseCase) ExecuteDeleteAll(ctx context.Context, req DeleteAllRequest) (*DeleteAllResponse, error) {
	// Get all sessions
	sessions, _, err := uc.sessionRepo.List(ctx, 1000, 0) // Get up to 1000 sessions
	if err != nil {
		uc.logger.ErrorWithError("failed to list sessions for deletion", err, nil)
		return nil, err
	}

	response := &DeleteAllResponse{}
	var errors []string

	// Delete each session
	for _, sess := range sessions {
		deleteReq := DeleteRequest{
			SessionID: sess.ID(),
			Force:     req.Force,
		}

		_, err := uc.Execute(ctx, deleteReq)
		if err != nil {
			response.FailedCount++
			errorMsg := fmt.Sprintf("Failed to delete session %s: %v", sess.ID().String(), err)
			errors = append(errors, errorMsg)
			uc.logger.ErrorWithError("failed to delete session in bulk operation", err, logger.Fields{
				"session_id": sess.ID().String(),
			})
		} else {
			response.DeletedCount++
		}
	}

	response.Errors = errors
	if response.FailedCount == 0 {
		response.Message = fmt.Sprintf("All %d sessions deleted successfully", response.DeletedCount)
	} else {
		response.Message = fmt.Sprintf("Deleted %d sessions, failed to delete %d sessions", response.DeletedCount, response.FailedCount)
	}

	uc.logger.InfoWithFields("bulk session deletion completed", logger.Fields{
		"deleted_count": response.DeletedCount,
		"failed_count":  response.FailedCount,
		"force":         req.Force,
	})

	return response, nil
}
