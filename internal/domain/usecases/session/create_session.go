package usecase

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
)

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	ProxyURL string `json:"proxy_url,omitempty" validate:"omitempty,url"`
}

// CreateSessionResponse represents the response after creating a session
type CreateSessionResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// CreateSessionUseCase handles the creation of new sessions
type CreateSessionUseCase struct {
	sessionRepo Repository
}

// NewCreateSessionUseCase creates a new instance of CreateSessionUseCase
func NewCreateSessionUseCase(sessionRepo Repository) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		sessionRepo: sessionRepo,
	}
}

// Execute creates a new session
func (uc *CreateSessionUseCase) Execute(ctx context.Context, req CreateSessionRequest) (*CreateSessionResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, err
	}

	// Check if session with same name already exists
	exists, err := uc.sessionRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		log.Error().Err(err).Str("name", req.Name).Msg("Failed to check session existence")
		return nil, err
	}

	if exists {
		return nil, ErrSessionAlreadyExists(req.Name)
	}

	// Create new session entity
	sess := NewSession(req.Name)

	// Set proxy URL if provided
	if req.ProxyURL != "" {
		if err := sess.SetProxyURL(req.ProxyURL); err != nil {
			return nil, err
		}
	}

	// Save session to repository
	if err := uc.sessionRepo.Create(ctx, sess); err != nil {
		log.Error().Err(err).Str("session_id", sess.ID().String()).Msg("Failed to create session")
		return nil, err
	}

	log.Info().
		Str("session_id", sess.ID().String()).
		Str("name", sess.Name()).
		Msg("Session created successfully")

	// Return response
	return &CreateSessionResponse{
		ID:        sess.ID().String(),
		Name:      sess.Name(),
		Status:    string(sess.Status()),
		CreatedAt: sess.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// validateRequest validates the create session request
func (uc *CreateSessionUseCase) validateRequest(req CreateSessionRequest) error {
	// Validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return ErrInvalidSessionName("name cannot be empty")
	}

	if len(name) > 255 {
		return ErrInvalidSessionName("name cannot exceed 255 characters")
	}

	// Validate proxy URL if provided
	if req.ProxyURL != "" {
		// Basic URL validation is handled by the domain entity
		// Additional validation can be added here if needed
	}

	return nil
}
