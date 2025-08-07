package repository

import (
	"context"
	"database/sql"
	"fmt"

	"wazmeow/internal/domain"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

// sessionRepository implements the domain.Repository interface
type sessionRepository struct {
	db *bun.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *bun.DB) domain.Repository {
	return &sessionRepository{db: db}
}

// Create stores a new session
func (r *sessionRepository) Create(ctx context.Context, sess *domain.Session) error {
	_, err := r.db.NewInsert().Model(sess).Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("session_id", sess.ID.String()).Msg("Failed to create session")
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Info().Str("session_id", sess.ID.String()).Str("name", sess.Name).Msg("Session created")
	return nil
}

// GetByID retrieves a session by its ID
func (r *sessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	session := new(domain.Session)
	err := r.db.NewSelect().
		Model(session).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound(id)
		}
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to get session by ID")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// GetByName retrieves a session by its name
func (r *sessionRepository) GetByName(ctx context.Context, name string) (*domain.Session, error) {
	session := new(domain.Session)
	err := r.db.NewSelect().
		Model(session).
		Where("name = ?", name).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewNotFoundError("Session", name)
		}
		log.Error().Err(err).Str("name", name).Msg("Failed to get session by name")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// GetAll retrieves all sessions
func (r *sessionRepository) GetAll(ctx context.Context) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.NewSelect().
		Model(&sessions).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get all sessions")
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return sessions, nil
}

// GetActive retrieves all active sessions
func (r *sessionRepository) GetActive(ctx context.Context) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.NewSelect().
		Model(&sessions).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get active sessions")
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return sessions, nil
}

// GetByStatus retrieves sessions by status
func (r *sessionRepository) GetByStatus(ctx context.Context, status domain.Status) ([]*domain.Session, error) {
	var sessions []*domain.Session
	err := r.db.NewSelect().
		Model(&sessions).
		Where("status = ?", status).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		log.Error().Err(err).Str("status", string(status)).Msg("Failed to get sessions by status")
		return nil, fmt.Errorf("failed to get sessions by status: %w", err)
	}

	return sessions, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(ctx context.Context, sess *domain.Session) error {
	result, err := r.db.NewUpdate().
		Model(sess).
		Where("id = ?", sess.ID).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Str("session_id", sess.ID.String()).Msg("Failed to update session")
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound(sess.ID)
	}

	log.Info().Str("session_id", sess.ID.String()).Msg("Session updated")
	return nil
}

// Delete removes a session by its ID
func (r *sessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	result, err := r.db.NewDelete().
		Model((*domain.Session)(nil)).
		Where("id = ?", id).
		Exec(ctx)

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

// ExistsByID checks if a session exists by ID
func (r *sessionRepository) ExistsByID(ctx context.Context, id domain.SessionID) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*domain.Session)(nil)).
		Where("id = ?", id).
		Count(ctx)

	if err != nil {
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to check session existence")
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByName checks if a session exists by name
func (r *sessionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*domain.Session)(nil)).
		Where("name = ?", name).
		Count(ctx)

	if err != nil {
		log.Error().Err(err).Str("name", name).Msg("Failed to check session existence by name")
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// Count returns the total number of sessions
func (r *sessionRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*domain.Session)(nil)).
		Count(ctx)

	if err != nil {
		log.Error().Err(err).Msg("Failed to count sessions")
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return int64(count), nil
}

// CountByStatus returns the number of sessions by status
func (r *sessionRepository) CountByStatus(ctx context.Context, status domain.Status) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*domain.Session)(nil)).
		Where("status = ?", status).
		Count(ctx)

	if err != nil {
		log.Error().Err(err).Str("status", string(status)).Msg("Failed to count sessions by status")
		return 0, fmt.Errorf("failed to count sessions by status: %w", err)
	}

	return int64(count), nil
}

// GetActiveCount returns the count of active sessions
func (r *sessionRepository) GetActiveCount(ctx context.Context) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*domain.Session)(nil)).
		Where("is_active = ?", true).
		Count(ctx)

	if err != nil {
		log.Error().Err(err).Msg("Failed to count active sessions")
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}

	return int64(count), nil
}

// List retrieves all sessions with optional filters
func (r *sessionRepository) List(ctx context.Context, filters map[string]any) ([]*domain.Session, error) {
	// For now, just return all sessions - filters can be implemented later
	return r.GetAll(ctx)
}

// UpdateStatus updates the status of a session
func (r *sessionRepository) UpdateStatus(ctx context.Context, id domain.SessionID, status domain.Status) error {
	// Use bun's query builder instead of raw SQL
	result, err := r.db.NewUpdate().
		Model((*domain.Session)(nil)).
		Set("status = ?", string(status)).
		Set("updated_at = NOW()").
		Where("id = ?", id.String()).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to update session status")
		return fmt.Errorf("failed to update session status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound(id)
	}

	log.Info().
		Str("session_id", id.String()).
		Str("status", string(status)).
		Int64("rows_affected", rowsAffected).
		Msg("Session status updated successfully")

	return nil
}

// SetWAJID sets the WhatsApp JID for a session
func (r *sessionRepository) SetWAJID(ctx context.Context, id domain.SessionID, wajid string) error {
	// Use bun's query builder instead of raw SQL
	result, err := r.db.NewUpdate().
		Model((*domain.Session)(nil)).
		Set("wa_jid = ?", wajid).
		Set("updated_at = NOW()").
		Where("id = ?", id.String()).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to set WAJID")
		return fmt.Errorf("failed to set WAJID: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound(id)
	}

	log.Info().
		Str("session_id", id.String()).
		Str("wa_jid", wajid).
		Int64("rows_affected", rowsAffected).
		Msg("WhatsApp JID updated successfully")

	return nil
}

// SetQRCode sets the QR code for a session
func (r *sessionRepository) SetQRCode(ctx context.Context, id domain.SessionID, qrCode string) error {
	// Use bun's query builder instead of raw SQL
	result, err := r.db.NewUpdate().
		Model((*domain.Session)(nil)).
		Set("qr_code = ?", qrCode).
		Set("updated_at = NOW()").
		Where("id = ?", id.String()).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to set QR code")
		return fmt.Errorf("failed to set QR code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrSessionNotFound(id)
	}

	log.Info().
		Str("session_id", id.String()).
		Int64("rows_affected", rowsAffected).
		Msg("QR code updated successfully")

	return nil
}

// ClearQRCode clears the QR code for a session
func (r *sessionRepository) ClearQRCode(ctx context.Context, id domain.SessionID) error {
	return r.SetQRCode(ctx, id, "")
}

// GetConnectedSessions retrieves all connected sessions
func (r *sessionRepository) GetConnectedSessions(ctx context.Context) ([]*domain.Session, error) {
	return r.GetByStatus(ctx, domain.StatusConnected)
}

// BulkUpdateStatus updates status for multiple sessions
func (r *sessionRepository) BulkUpdateStatus(ctx context.Context, ids []domain.SessionID, status domain.Status) error {
	if len(ids) == 0 {
		return nil
	}

	// Convert SessionIDs to strings for the query
	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	// Use bun's query builder for bulk update
	result, err := r.db.NewUpdate().
		Model((*domain.Session)(nil)).
		Set("status = ?", string(status)).
		Set("updated_at = NOW()").
		Where("id IN (?)", bun.In(stringIDs)).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Interface("session_ids", stringIDs).Msg("Failed to bulk update session status")
		return fmt.Errorf("failed to bulk update session status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get rows affected for bulk update")
	} else {
		log.Info().
			Interface("session_ids", stringIDs).
			Str("status", string(status)).
			Int64("rows_affected", rowsAffected).
			Msg("Bulk session status updated successfully")
	}

	return nil
}
