package usecase

import (
	"context"

	"wazmeow/internal/domain/session"

	"github.com/rs/zerolog/log"
)

// ListSessionsRequest represents the request to list sessions
type ListSessionsRequest struct {
	ActiveOnly bool   `json:"active_only,omitempty"`
	Status     string `json:"status,omitempty"`
}

// SessionInfo represents session information in the response
type SessionInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	WAJID     string `json:"wa_jid,omitempty"`
	ProxyURL  string `json:"proxy_url,omitempty"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListSessionsResponse represents the response for listing sessions
type ListSessionsResponse struct {
	Sessions []SessionInfo `json:"sessions"`
	Total    int           `json:"total"`
}

// ListSessionsUseCase handles listing sessions
type ListSessionsUseCase struct {
	sessionRepo session.Repository
}

// NewListSessionsUseCase creates a new instance of ListSessionsUseCase
func NewListSessionsUseCase(sessionRepo session.Repository) *ListSessionsUseCase {
	return &ListSessionsUseCase{
		sessionRepo: sessionRepo,
	}
}

// Execute lists sessions based on the request criteria
func (uc *ListSessionsUseCase) Execute(ctx context.Context, req ListSessionsRequest) (*ListSessionsResponse, error) {
	var sessions []*session.Session
	var err error

	// Determine which sessions to retrieve based on request
	switch {
	case req.ActiveOnly:
		sessions, err = uc.sessionRepo.GetActive(ctx)
	case req.Status != "":
		status := session.Status(req.Status)
		if !status.IsValid() {
			return nil, session.ErrInvalidStatus(req.Status)
		}
		sessions, err = uc.sessionRepo.GetByStatus(ctx, status)
	default:
		sessions, err = uc.sessionRepo.GetAll(ctx)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve sessions")
		return nil, err
	}

	// Convert domain entities to response format
	sessionInfos := make([]SessionInfo, len(sessions))
	for i, sess := range sessions {
		sessionInfos[i] = SessionInfo{
			ID:        sess.ID().String(),
			Name:      sess.Name(),
			Status:    string(sess.Status()),
			WAJID:     sess.WAJID(),
			ProxyURL:  sess.ProxyURL(),
			IsActive:  sess.IsActive(),
			CreatedAt: sess.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: sess.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	log.Info().
		Int("count", len(sessions)).
		Bool("active_only", req.ActiveOnly).
		Str("status_filter", req.Status).
		Msg("Sessions retrieved successfully")

	return &ListSessionsResponse{
		Sessions: sessionInfos,
		Total:    len(sessionInfos),
	}, nil
}
