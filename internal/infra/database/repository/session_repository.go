package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wazmeow/internal/domain"
	"wazmeow/internal/infra/database/models"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// sessionRepository implements the domain.Repository interface
type sessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB) domain.Repository {
	return &sessionRepository{db: db}
}

// Create stores a new session
func (r *sessionRepository) Create(ctx context.Context, sess *domain.Session) error {
	model := models.NewSessionModelFromEntity(sess)

	query := `
		INSERT INTO sessions (id, name, status, wa_jid, qr_code, proxy_url, is_active, created_at, updated_at)
		VALUES (:id, :name, :status, :wa_jid, :qr_code, :proxy_url, :is_active, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		log.Error().Err(err).Str("session_id", sess.ID().String()).Msg("Failed to create session")
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Info().Str("session_id", sess.ID().String()).Str("name", sess.Name()).Msg("Session created")
	return nil
}

// GetByID retrieves a session by its ID
func (r *sessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	var model models.SessionModel
	query := `SELECT id, name, status, wa_jid, qr_code, proxy_url, is_active, created_at, updated_at 
			  FROM sessions WHERE id = $1`

	err := r.db.GetContext(ctx, &model, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound(id)
		}
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to get session by ID")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return model.ToEntity(), nil
}

// GetByName retrieves a session by its name
func (r *sessionRepository) GetByName(ctx context.Context, name string) (*domain.Session, error) {
	var model models.SessionModel
	query := `SELECT id, name, status, wa_jid, qr_code, proxy_url, is_active, created_at, updated_at 
			  FROM sessions WHERE name = $1`

	err := r.db.GetContext(ctx, &model, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewNotFoundError(fmt.Sprintf("session with name '%s' not found", name))
		}
		log.Error().Err(err).Str("name", name).Msg("Failed to get session by name")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return model.ToEntity(), nil
}

// GetAll retrieves all sessions
func (r *sessionRepository) GetAll(ctx context.Context) ([]*domain.Session, error) {
	var models []models.SessionModel
	query := `SELECT id, name, status, wa_jid, qr_code, proxy_url, is_active, created_at, updated_at 
			  FROM sessions ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all sessions")
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	sessions := make([]*domain.Session, len(models))
	for i, model := range models {
		sessions[i] = model.ToEntity()
	}

	return sessions, nil
}

// GetActive retrieves all active sessions
func (r *sessionRepository) GetActive(ctx context.Context) ([]*domain.Session, error) {
	var models []models.SessionModel
	query := `SELECT id, name, status, wa_jid, qr_code, proxy_url, is_active, created_at, updated_at 
			  FROM sessions WHERE is_active = true ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get active sessions")
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	sessions := make([]*domain.Session, len(models))
	for i, model := range models {
		sessions[i] = model.ToEntity()
	}

	return sessions, nil
}

// GetByStatus retrieves sessions by status
func (r *sessionRepository) GetByStatus(ctx context.Context, status domain.Status) ([]*domain.Session, error) {
	var models []models.SessionModel
	query := `SELECT id, name, status, wa_jid, qr_code, proxy_url, is_active, created_at, updated_at 
			  FROM sessions WHERE status = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &models, query, string(status))
	if err != nil {
		log.Error().Err(err).Str("status", string(status)).Msg("Failed to get sessions by status")
		return nil, fmt.Errorf("failed to get sessions by status: %w", err)
	}

	sessions := make([]*domain.Session, len(models))
	for i, model := range models {
		sessions[i] = model.ToEntity()
	}

	return sessions, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(ctx context.Context, sess *domain.Session) error {
	model := models.NewSessionModelFromEntity(sess)

	query := `
		UPDATE sessions 
		SET name = :name, status = :status, wa_jid = :wa_jid, qr_code = :qr_code, 
			proxy_url = :proxy_url, is_active = :is_active, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		log.Error().Err(err).Str("session_id", sess.ID().String()).Msg("Failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound(sess.ID())
	}

	log.Info().Str("session_id", sess.ID().String()).Msg("Session updated")
	return nil
}

// Delete removes a session by its ID
func (r *sessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound(id)
	}

	log.Info().Str("session_id", id.String()).Msg("Session deleted")
	return nil
}

// Exists checks if a session exists by ID
func (r *sessionRepository) Exists(ctx context.Context, id domain.SessionID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sessions WHERE id = $1`

	err := r.db.GetContext(ctx, &count, query, id.String())
	if err != nil {
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to check session existence")
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if a session exists by name
func (r *sessionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM sessions WHERE name = $1`

	err := r.db.GetContext(ctx, &count, query, name)
	if err != nil {
		log.Error().Err(err).Str("name", name).Msg("Failed to check session existence by name")
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// Count returns the total number of sessions
func (r *sessionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM sessions`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count sessions")
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return count, nil
}

// CountByStatus returns the number of sessions by status
func (r *sessionRepository) CountByStatus(ctx context.Context, status domain.Status) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM sessions WHERE status = $1`

	err := r.db.GetContext(ctx, &count, query, string(status))
	if err != nil {
		log.Error().Err(err).Str("status", string(status)).Msg("Failed to count sessions by status")
		return 0, fmt.Errorf("failed to count sessions by status: %w", err)
	}

	return count, nil
}
