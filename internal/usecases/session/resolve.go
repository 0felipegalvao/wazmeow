package session

import (
	"context"
	"fmt"

	"wazmeow/internal/domain/session"
	"wazmeow/pkg/logger"
)

// ResolveUseCase handles resolving sessions by flexible identifier (ID or name)
type ResolveUseCase struct {
	repo   session.Repository
	logger logger.Logger
}

// NewResolveUseCase creates a new resolve use case
func NewResolveUseCase(repo session.Repository, logger logger.Logger) *ResolveUseCase {
	return &ResolveUseCase{
		repo:   repo,
		logger: logger,
	}
}

// ResolveRequest represents the request to resolve a session
type ResolveRequest struct {
	Identifier session.SessionIdentifier `json:"identifier"`
}

// ResolveResponse represents the response from resolving a session
type ResolveResponse struct {
	Session        *session.Session `json:"session"`
	IdentifierType string           `json:"identifier_type"`
}

// Execute resolves a session by its flexible identifier
func (uc *ResolveUseCase) Execute(ctx context.Context, req ResolveRequest) (*ResolveResponse, error) {
	// Validate the identifier first
	if err := req.Identifier.Validate(); err != nil {
		uc.logger.ErrorWithError("invalid session identifier", err, logger.Fields{
			"identifier": req.Identifier.String(),
		})
		return nil, err
	}

	// Log the resolution attempt
	uc.logger.InfoWithFields("resolving session", logger.Fields{
		"identifier":      req.Identifier.String(),
		"identifier_type": req.Identifier.Type().String(),
	})

	var sess *session.Session
	var err error

	// Resolve based on identifier type
	if req.Identifier.IsID() {
		// Resolve by SessionID
		sessionID, convErr := req.Identifier.ToSessionID()
		if convErr != nil {
			uc.logger.ErrorWithError("failed to convert identifier to session ID", convErr, logger.Fields{
				"identifier": req.Identifier.String(),
			})
			return nil, fmt.Errorf("invalid session ID format: %w", convErr)
		}

		sess, err = uc.repo.GetByID(ctx, sessionID)
		if err != nil {
			if err == session.ErrSessionNotFound {
				uc.logger.WarnWithFields("session not found by ID", logger.Fields{
					"session_id": sessionID.String(),
				})
				return nil, fmt.Errorf("session with ID '%s' not found", sessionID.String())
			}
			uc.logger.ErrorWithError("failed to get session by ID", err, logger.Fields{
				"session_id": sessionID.String(),
			})
			return nil, fmt.Errorf("failed to retrieve session by ID: %w", err)
		}
	} else if req.Identifier.IsName() {
		// Resolve by SessionName
		sessionName, convErr := req.Identifier.ToSessionName()
		if convErr != nil {
			uc.logger.ErrorWithError("failed to convert identifier to session name", convErr, logger.Fields{
				"identifier": req.Identifier.String(),
			})
			return nil, fmt.Errorf("invalid session name format: %w", convErr)
		}

		sess, err = uc.repo.GetByName(ctx, sessionName.String())
		if err != nil {
			if err == session.ErrSessionNotFound {
				uc.logger.WarnWithFields("session not found by name", logger.Fields{
					"session_name": sessionName.String(),
				})
				return nil, fmt.Errorf("session with name '%s' not found", sessionName.String())
			}
			uc.logger.ErrorWithError("failed to get session by name", err, logger.Fields{
				"session_name": sessionName.String(),
			})
			return nil, fmt.Errorf("failed to retrieve session by name: %w", err)
		}
	} else {
		// This should not happen if SessionIdentifier is properly implemented
		uc.logger.ErrorWithFields("unknown identifier type", logger.Fields{
			"identifier":      req.Identifier.String(),
			"identifier_type": req.Identifier.Type().String(),
		})
		return nil, fmt.Errorf("unsupported identifier type: %s", req.Identifier.Type().String())
	}

	// Log successful resolution
	uc.logger.InfoWithFields("session resolved successfully", logger.Fields{
		"session_id":      sess.ID().String(),
		"session_name":    sess.Name(),
		"identifier":      req.Identifier.String(),
		"identifier_type": req.Identifier.Type().String(),
	})

	return &ResolveResponse{
		Session:        sess,
		IdentifierType: req.Identifier.Type().String(),
	}, nil
}
