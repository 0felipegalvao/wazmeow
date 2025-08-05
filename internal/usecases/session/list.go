package session

import (
	"context"

	"wazmeow/internal/domain/session"
	"wazmeow/pkg/logger"
)

// ListUseCase handles listing sessions
type ListUseCase struct {
	repo   session.Repository
	logger logger.Logger
}

// NewListUseCase creates a new list sessions use case
func NewListUseCase(repo session.Repository, logger logger.Logger) *ListUseCase {
	return &ListUseCase{
		repo:   repo,
		logger: logger,
	}
}

// ListRequest represents the request to list sessions
type ListRequest struct {
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}

// ListResponse represents the response from listing sessions
type ListResponse struct {
	Sessions []*session.Session `json:"sessions"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

// Execute lists sessions with pagination
func (uc *ListUseCase) Execute(ctx context.Context, req ListRequest) (*ListResponse, error) {
	// Set default values
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get sessions from repository
	sessions, total, err := uc.repo.List(ctx, req.Limit, req.Offset)
	if err != nil {
		uc.logger.ErrorWithError("failed to list sessions", err, logger.Fields{
			"limit":  req.Limit,
			"offset": req.Offset,
		})
		return nil, err
	}

	uc.logger.InfoWithFields("sessions listed successfully", logger.Fields{
		"count":  len(sessions),
		"total":  total,
		"limit":  req.Limit,
		"offset": req.Offset,
	})

	return &ListResponse{
		Sessions: sessions,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

// ListByStatusRequest represents the request to list sessions by status
type ListByStatusRequest struct {
	Status session.Status `json:"status"`
	Limit  int            `json:"limit" validate:"min=1,max=100"`
	Offset int            `json:"offset" validate:"min=0"`
}

// ExecuteByStatus lists sessions filtered by status
func (uc *ListUseCase) ExecuteByStatus(ctx context.Context, req ListByStatusRequest) (*ListResponse, error) {
	// Set default values
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get sessions by status from repository
	sessions, total, err := uc.repo.GetByStatus(ctx, req.Status, req.Limit, req.Offset)
	if err != nil {
		uc.logger.ErrorWithError("failed to list sessions by status", err, logger.Fields{
			"status": req.Status.String(),
			"limit":  req.Limit,
			"offset": req.Offset,
		})
		return nil, err
	}

	uc.logger.InfoWithFields("sessions listed by status successfully", logger.Fields{
		"status": req.Status.String(),
		"count":  len(sessions),
		"total":  total,
		"limit":  req.Limit,
		"offset": req.Offset,
	})

	return &ListResponse{
		Sessions: sessions,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
	}, nil
}

// GetActiveCountRequest represents the request to get active session count
type GetActiveCountRequest struct{}

// GetActiveCountResponse represents the response with active session count
type GetActiveCountResponse struct {
	Count int `json:"count"`
}

// ExecuteGetActiveCount gets the count of active sessions
func (uc *ListUseCase) ExecuteGetActiveCount(ctx context.Context, req GetActiveCountRequest) (*GetActiveCountResponse, error) {
	count, err := uc.repo.GetActiveCount(ctx)
	if err != nil {
		uc.logger.ErrorWithError("failed to get active session count", err, nil)
		return nil, err
	}

	uc.logger.InfoWithFields("active session count retrieved", logger.Fields{
		"count": count,
	})

	return &GetActiveCountResponse{
		Count: count,
	}, nil
}
