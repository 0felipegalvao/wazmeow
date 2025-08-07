package domain

import "context"

// Repository defines the interface for session persistence
type Repository interface {
	// Create stores a new session
	Create(ctx context.Context, session *Session) error

	// GetByID retrieves a session by its ID
	GetByID(ctx context.Context, id SessionID) (*Session, error)

	// GetByName retrieves a session by its name
	GetByName(ctx context.Context, name string) (*Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *Session) error

	// Delete removes a session
	Delete(ctx context.Context, id SessionID) error

	// List retrieves all sessions with optional filters
	List(ctx context.Context, filters map[string]interface{}) ([]*Session, error)

	// ExistsByID checks if a session exists by ID
	ExistsByID(ctx context.Context, id SessionID) (bool, error)

	// ExistsByName checks if a session exists by name
	ExistsByName(ctx context.Context, name string) (bool, error)

	// GetActiveCount returns the count of active sessions
	GetActiveCount(ctx context.Context) (int64, error)

	// GetByStatus retrieves sessions by status
	GetByStatus(ctx context.Context, status Status) ([]*Session, error)

	// UpdateStatus updates the status of a session
	UpdateStatus(ctx context.Context, id SessionID, status Status) error

	// SetWAJID sets the WhatsApp JID for a session
	SetWAJID(ctx context.Context, id SessionID, wajid string) error

	// SetQRCode sets the QR code for a session
	SetQRCode(ctx context.Context, id SessionID, qrCode string) error

	// ClearQRCode clears the QR code for a session
	ClearQRCode(ctx context.Context, id SessionID) error

	// GetConnectedSessions retrieves all connected sessions
	GetConnectedSessions(ctx context.Context) ([]*Session, error)

	// BulkUpdateStatus updates status for multiple sessions
	BulkUpdateStatus(ctx context.Context, ids []SessionID, status Status) error
}
