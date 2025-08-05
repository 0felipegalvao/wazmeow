package session

import "context"

// Repository defines the interface for session persistence operations
type Repository interface {
	// Create stores a new session in the repository
	Create(ctx context.Context, session *Session) error

	// GetByID retrieves a session by its ID
	GetByID(ctx context.Context, id SessionID) (*Session, error)

	// GetByName retrieves a session by its name
	GetByName(ctx context.Context, name string) (*Session, error)

	// List retrieves sessions with pagination
	List(ctx context.Context, limit, offset int) ([]*Session, int, error)

	// Update updates an existing session
	Update(ctx context.Context, session *Session) error

	// Delete removes a session from the repository
	Delete(ctx context.Context, id SessionID) error

	// UpdateStatus updates only the status of a session
	UpdateStatus(ctx context.Context, id SessionID, status Status) error

	// GetActiveCount returns the number of active sessions
	GetActiveCount(ctx context.Context) (int, error)

	// GetByStatus retrieves sessions by their status
	GetByStatus(ctx context.Context, status Status, limit, offset int) ([]*Session, int, error)

	// Exists checks if a session with the given ID exists
	Exists(ctx context.Context, id SessionID) (bool, error)

	// ExistsByName checks if a session with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// ListFilter represents filters for listing sessions
type ListFilter struct {
	Status   *Status
	IsActive *bool
	Search   string
}

// ListOptions represents options for listing sessions
type ListOptions struct {
	Limit  int
	Offset int
	Sort   string
	Order  string
}

// RepositoryWithFilters extends Repository with advanced filtering capabilities
type RepositoryWithFilters interface {
	Repository

	// ListWithFilter retrieves sessions with advanced filtering
	ListWithFilter(ctx context.Context, filter ListFilter, options ListOptions) ([]*Session, int, error)

	// CountWithFilter counts sessions matching the filter
	CountWithFilter(ctx context.Context, filter ListFilter) (int, error)
}
