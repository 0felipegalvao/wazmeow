package session

import (
	"context"

	"wazmeow/internal/domain/session"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// CreateUseCase handles session creation
type CreateUseCase struct {
	repo      session.Repository
	logger    logger.Logger
	validator validator.Validator
}

// NewCreateUseCase creates a new create session use case
func NewCreateUseCase(repo session.Repository, logger logger.Logger, validator validator.Validator) *CreateUseCase {
	return &CreateUseCase{
		repo:      repo,
		logger:    logger,
		validator: validator,
	}
}

// CreateRequest represents the request to create a session
type CreateRequest struct {
	Name string `json:"name" validate:"required,session_name"`
}

// CreateResponse represents the response from creating a session
type CreateResponse struct {
	Session *session.Session `json:"session"`
}

// Execute creates a new session
func (uc *CreateUseCase) Execute(ctx context.Context, req CreateRequest) (*CreateResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for create session", err, logger.Fields{
			"name": req.Name,
		})
		return nil, err
	}

	// Check if session with same name already exists
	existing, err := uc.repo.GetByName(ctx, req.Name)
	if err != nil && err != session.ErrSessionNotFound {
		uc.logger.ErrorWithError("failed to check existing session", err, logger.Fields{
			"name": req.Name,
		})
		return nil, err
	}

	if existing != nil {
		uc.logger.WarnWithFields("session with name already exists", logger.Fields{
			"name":       req.Name,
			"session_id": existing.ID().String(),
		})
		return nil, session.ErrSessionAlreadyExists
	}

	// Create new session
	sess := session.NewSession(req.Name)

	// Validate session entity
	if err := sess.Validate(); err != nil {
		uc.logger.ErrorWithError("session validation failed", err, logger.Fields{
			"name":       req.Name,
			"session_id": sess.ID().String(),
		})
		return nil, err
	}

	// Save to repository
	if err := uc.repo.Create(ctx, sess); err != nil {
		uc.logger.ErrorWithError("failed to create session", err, logger.Fields{
			"name":       req.Name,
			"session_id": sess.ID().String(),
		})
		return nil, err
	}

	uc.logger.InfoWithFields("session created successfully", logger.Fields{
		"name":       sess.Name(),
		"session_id": sess.ID().String(),
		"status":     sess.Status().String(),
	})

	return &CreateResponse{
		Session: sess,
	}, nil
}
