package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/infra/database"
	"wazmeow/pkg/logger"
)

// SessionRepository implements session.Repository using Bun ORM (supports SQLite, PostgreSQL, etc.)
type SessionRepository struct {
	db     *bun.DB
	logger logger.Logger
}

// NewSessionRepository creates a new session repository using Bun ORM
func NewSessionRepository(db *bun.DB, logger logger.Logger) session.Repository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create stores a new session in the repository
func (r *SessionRepository) Create(ctx context.Context, sess *session.Session) error {
	model := database.ToWazMeowSessionModel(sess)

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to create session", err, logger.Fields{
			"session_id": sess.ID().String(),
			"name":       sess.Name(),
		})
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.InfoWithFields("session created", logger.Fields{
		"session_id": sess.ID().String(),
		"name":       sess.Name(),
	})

	return nil
}

// GetByID retrieves a session by its ID
func (r *SessionRepository) GetByID(ctx context.Context, id session.SessionID) (*session.Session, error) {
	var model database.WazMeowSessionModel

	err := r.db.NewSelect().
		Model(&model).
		Where("id = ?", id.String()).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, session.ErrSessionNotFound
		}
		r.logger.ErrorWithError("failed to get session by ID", err, logger.Fields{
			"session_id": id.String(),
		})
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	sess, err := database.FromWazMeowSessionModel(&model)
	if err != nil {
		r.logger.ErrorWithError("failed to convert session model", err, logger.Fields{
			"session_id": id.String(),
		})
		return nil, fmt.Errorf("failed to convert session model: %w", err)
	}

	return sess, nil
}

// GetByName retrieves a session by its name
func (r *SessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	var model database.WazMeowSessionModel

	err := r.db.NewSelect().
		Model(&model).
		Where("name = ?", name).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, session.ErrSessionNotFound
		}
		r.logger.ErrorWithError("failed to get session by name", err, logger.Fields{
			"name": name,
		})
		return nil, fmt.Errorf("failed to get session by name: %w", err)
	}

	sess, err := database.FromWazMeowSessionModel(&model)
	if err != nil {
		r.logger.ErrorWithError("failed to convert session model", err, logger.Fields{
			"name": name,
		})
		return nil, fmt.Errorf("failed to convert session model: %w", err)
	}

	return sess, nil
}

// List retrieves sessions with pagination
func (r *SessionRepository) List(ctx context.Context, limit, offset int) ([]*session.Session, int, error) {
	var models []database.WazMeowSessionModel

	// Get sessions with pagination
	err := r.db.NewSelect().
		Model(&models).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to list sessions", err, logger.Fields{
			"limit":  limit,
			"offset": offset,
		})
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Get total count
	total, err := r.db.NewSelect().
		Model((*database.WazMeowSessionModel)(nil)).
		Count(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to count sessions", err, nil)
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	// Convert models to domain entities
	sessions := make([]*session.Session, 0, len(models))
	for _, model := range models {
		sess, err := database.FromWazMeowSessionModel(&model)
		if err != nil {
			r.logger.ErrorWithError("failed to convert session model", err, logger.Fields{
				"session_id": model.ID,
			})
			continue // Skip invalid sessions
		}
		sessions = append(sessions, sess)
	}

	return sessions, total, nil
}

// Update updates an existing session
func (r *SessionRepository) Update(ctx context.Context, sess *session.Session) error {
	model := database.ToWazMeowSessionModel(sess)

	result, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", sess.ID().String()).
		Exec(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to update session", err, logger.Fields{
			"session_id": sess.ID().String(),
		})
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return session.ErrSessionNotFound
	}

	r.logger.InfoWithFields("session updated", logger.Fields{
		"session_id": sess.ID().String(),
		"name":       sess.Name(),
		"status":     sess.Status().String(),
	})

	return nil
}

// Delete removes a session from the repository
func (r *SessionRepository) Delete(ctx context.Context, id session.SessionID) error {
	result, err := r.db.NewDelete().
		Model((*database.WazMeowSessionModel)(nil)).
		Where("id = ?", id.String()).
		Exec(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to delete session", err, logger.Fields{
			"session_id": id.String(),
		})
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return session.ErrSessionNotFound
	}

	r.logger.InfoWithFields("session deleted", logger.Fields{
		"session_id": id.String(),
	})

	return nil
}

// UpdateStatus updates only the status of a session
func (r *SessionRepository) UpdateStatus(ctx context.Context, id session.SessionID, status session.Status) error {
	result, err := r.db.NewUpdate().
		Model((*database.WazMeowSessionModel)(nil)).
		Set("status = ?", status.String()).
		Set("updated_at = CURRENT_TIMESTAMP").
		Where("id = ?", id.String()).
		Exec(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to update session status", err, logger.Fields{
			"session_id": id.String(),
			"status":     status.String(),
		})
		return fmt.Errorf("failed to update session status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return session.ErrSessionNotFound
	}

	r.logger.InfoWithFields("session status updated", logger.Fields{
		"session_id": id.String(),
		"status":     status.String(),
	})

	return nil
}

// GetActiveCount returns the number of active sessions
func (r *SessionRepository) GetActiveCount(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().
		Model((*database.WazMeowSessionModel)(nil)).
		Where("is_active = ?", true).
		Count(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to get active session count", err, nil)
		return 0, fmt.Errorf("failed to get active session count: %w", err)
	}

	return count, nil
}

// GetByStatus retrieves sessions by their status
func (r *SessionRepository) GetByStatus(ctx context.Context, status session.Status, limit, offset int) ([]*session.Session, int, error) {
	var models []database.WazMeowSessionModel

	// Get sessions with pagination and status filter
	err := r.db.NewSelect().
		Model(&models).
		Where("status = ?", status.String()).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to get sessions by status", err, logger.Fields{
			"status": status.String(),
			"limit":  limit,
			"offset": offset,
		})
		return nil, 0, fmt.Errorf("failed to get sessions by status: %w", err)
	}

	// Get total count for this status
	total, err := r.db.NewSelect().
		Model((*database.WazMeowSessionModel)(nil)).
		Where("status = ?", status.String()).
		Count(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to count sessions by status", err, logger.Fields{
			"status": status.String(),
		})
		return nil, 0, fmt.Errorf("failed to count sessions by status: %w", err)
	}

	// Convert models to domain entities
	sessions := make([]*session.Session, 0, len(models))
	for _, model := range models {
		sess, err := database.FromWazMeowSessionModel(&model)
		if err != nil {
			r.logger.ErrorWithError("failed to convert session model", err, logger.Fields{
				"session_id": model.ID,
			})
			continue // Skip invalid sessions
		}
		sessions = append(sessions, sess)
	}

	return sessions, total, nil
}

// Exists checks if a session with the given ID exists
func (r *SessionRepository) Exists(ctx context.Context, id session.SessionID) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*database.WazMeowSessionModel)(nil)).
		Where("id = ?", id.String()).
		Count(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to check session existence", err, logger.Fields{
			"session_id": id.String(),
		})
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if a session with the given name exists
func (r *SessionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*database.WazMeowSessionModel)(nil)).
		Where("name = ?", name).
		Count(ctx)

	if err != nil {
		r.logger.ErrorWithError("failed to check session existence by name", err, logger.Fields{
			"name": name,
		})
		return false, fmt.Errorf("failed to check session existence by name: %w", err)
	}

	return count > 0, nil
}
