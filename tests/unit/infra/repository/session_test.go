package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/infra/database/migrations"
	"wazmeow/internal/infra/repository"
	"wazmeow/pkg/logger"
)

// TestDatabase represents a test database configuration
type TestDatabase struct {
	Name   string
	Driver string
	DSN    string
	DB     *bun.DB
}

func setupTestDB(t *testing.T) *bun.DB {
	// Create in-memory SQLite database for testing
	sqldb, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create bun DB instance
	db := bun.NewDB(sqldb, sqlitedialect.New())

	// Use migrations to create schema
	setupSchema(t, db)

	return db
}

// setupTestDatabases creates test databases for SQLite and PostgreSQL
func setupTestDatabases(t *testing.T) []TestDatabase {
	var databases []TestDatabase

	// SQLite in-memory database
	sqliteDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	bunSQLite := bun.NewDB(sqliteDB, sqlitedialect.New())
	databases = append(databases, TestDatabase{
		Name:   "SQLite",
		Driver: "sqlite3",
		DSN:    ":memory:",
		DB:     bunSQLite,
	})

	// PostgreSQL database (if available)
	if pgDSN := os.Getenv("TEST_POSTGRES_DSN"); pgDSN != "" {
		pgDB, err := sql.Open("postgres", pgDSN)
		if err == nil && pgDB.Ping() == nil {
			bunPG := bun.NewDB(pgDB, pgdialect.New())
			databases = append(databases, TestDatabase{
				Name:   "PostgreSQL",
				Driver: "postgres",
				DSN:    pgDSN,
				DB:     bunPG,
			})
		}
	}

	return databases
}

// setupSchema creates the necessary tables for testing
func setupSchema(t *testing.T, db *bun.DB) {
	nullLogger := &NullLogger{}
	migrator := migrations.NewMigrator(db, nullLogger)

	ctx := context.Background()
	err := migrator.Migrate(ctx)
	require.NoError(t, err, "Failed to run migrations")
}

// NullLogger for repository tests - implements logger.Logger interface
type NullLogger struct{}

func (n *NullLogger) Debug(msg string)                                           {}
func (n *NullLogger) Info(msg string)                                            {}
func (n *NullLogger) Warn(msg string)                                            {}
func (n *NullLogger) Error(msg string)                                           {}
func (n *NullLogger) Fatal(msg string)                                           {}
func (n *NullLogger) DebugWithFields(msg string, fields logger.Fields)           {}
func (n *NullLogger) InfoWithFields(msg string, fields logger.Fields)            {}
func (n *NullLogger) WarnWithFields(msg string, fields logger.Fields)            {}
func (n *NullLogger) ErrorWithFields(msg string, fields logger.Fields)           {}
func (n *NullLogger) FatalWithFields(msg string, fields logger.Fields)           {}
func (n *NullLogger) DebugWithError(msg string, err error, fields logger.Fields) {}
func (n *NullLogger) InfoWithError(msg string, err error, fields logger.Fields)  {}
func (n *NullLogger) WarnWithError(msg string, err error, fields logger.Fields)  {}
func (n *NullLogger) ErrorWithError(msg string, err error, fields logger.Fields) {}
func (n *NullLogger) FatalWithError(msg string, err error, fields logger.Fields) {}
func (n *NullLogger) WithContext(ctx context.Context) logger.Logger              { return n }
func (n *NullLogger) WithFields(fields logger.Fields) logger.Logger              { return n }
func (n *NullLogger) WithField(key string, value interface{}) logger.Logger      { return n }
func (n *NullLogger) WithError(err error) logger.Logger                          { return n }
func (n *NullLogger) SetLevel(level logger.Level)                                {}
func (n *NullLogger) GetLevel() logger.Level                                     { return logger.InfoLevel }
func (n *NullLogger) SetOutput(output io.Writer)                                 {}
func (n *NullLogger) IsDebugEnabled() bool                                       { return false }
func (n *NullLogger) IsInfoEnabled() bool                                        { return false }
func (n *NullLogger) IsWarnEnabled() bool                                        { return false }
func (n *NullLogger) IsErrorEnabled() bool                                       { return false }

func TestSessionRepository_Create(t *testing.T) {
	t.Run("should create session successfully", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		sess := session.NewSession("test-session")
		ctx := context.Background()

		// Act
		err := repo.Create(ctx, sess)

		// Assert
		assert.NoError(t, err)

		// Verify session was created in database
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM wazmeow_sessions WHERE id = ?", sess.ID().String()).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("should fail when session with same name exists", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		sess1 := session.NewSession("duplicate-session")
		sess2 := session.NewSession("duplicate-session")
		ctx := context.Background()

		// Act - Create first session
		err := repo.Create(ctx, sess1)
		require.NoError(t, err)

		// Act - Try to create second session with same name
		err = repo.Create(ctx, sess2)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed")
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		sess := session.NewSession("context-test")

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		err := repo.Create(ctx, sess)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestSessionRepository_GetByID(t *testing.T) {
	t.Run("should get session by ID successfully", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		originalSess := session.NewSession("get-by-id-test")
		ctx := context.Background()

		// Create session first
		err := repo.Create(ctx, originalSess)
		require.NoError(t, err)

		// Act
		retrievedSess, err := repo.GetByID(ctx, originalSess.ID())

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, retrievedSess)
		assert.Equal(t, originalSess.ID(), retrievedSess.ID())
		assert.Equal(t, originalSess.Name(), retrievedSess.Name())
		assert.Equal(t, originalSess.Status(), retrievedSess.Status())
		assert.Equal(t, originalSess.IsActive(), retrievedSess.IsActive())
	})

	t.Run("should return error when session not found", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		nonExistentID := session.NewSessionID()
		ctx := context.Background()

		// Act
		retrievedSess, err := repo.GetByID(ctx, nonExistentID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionNotFound, err)
		assert.Nil(t, retrievedSess)
	})

	t.Run("should get connected session correctly", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		originalSess := session.NewSession("connected-test")
		err := originalSess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)
		ctx := context.Background()

		// Create session first
		err = repo.Create(ctx, originalSess)
		require.NoError(t, err)

		// Act
		retrievedSess, err := repo.GetByID(ctx, originalSess.ID())

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, retrievedSess)
		assert.Equal(t, session.StatusConnected, retrievedSess.Status())
		assert.Equal(t, "test@s.whatsapp.net", retrievedSess.WaJID())
		assert.True(t, retrievedSess.IsActive())
	})
}

func TestSessionRepository_GetByName(t *testing.T) {
	t.Run("should get session by name successfully", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		originalSess := session.NewSession("get-by-name-test")
		ctx := context.Background()

		// Create session first
		err := repo.Create(ctx, originalSess)
		require.NoError(t, err)

		// Act
		retrievedSess, err := repo.GetByName(ctx, "get-by-name-test")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, retrievedSess)
		assert.Equal(t, originalSess.ID(), retrievedSess.ID())
		assert.Equal(t, originalSess.Name(), retrievedSess.Name())
	})

	t.Run("should return error when session not found by name", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Act
		retrievedSess, err := repo.GetByName(ctx, "non-existent-session")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionNotFound, err)
		assert.Nil(t, retrievedSess)
	})
}

func TestSessionRepository_Update(t *testing.T) {
	t.Run("should update session successfully", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		sess := session.NewSession("update-test")
		ctx := context.Background()

		// Create session first
		err := repo.Create(ctx, sess)
		require.NoError(t, err)

		// Modify session
		err = sess.Connect("updated@s.whatsapp.net")
		require.NoError(t, err)
		sess.SetQRCode("updated-qr-code")

		// Act
		err = repo.Update(ctx, sess)

		// Assert
		assert.NoError(t, err)

		// Verify update in database
		retrievedSess, err := repo.GetByID(ctx, sess.ID())
		require.NoError(t, err)
		assert.Equal(t, session.StatusConnected, retrievedSess.Status())
		assert.Equal(t, "updated@s.whatsapp.net", retrievedSess.WaJID())
		assert.Equal(t, "updated-qr-code", retrievedSess.QRCode())
		assert.True(t, retrievedSess.IsActive())
	})

	t.Run("should fail when updating non-existent session", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		sess := session.NewSession("non-existent")
		ctx := context.Background()

		// Act
		err := repo.Update(ctx, sess)

		// Assert
		assert.Error(t, err)
	})

	t.Run("should update timestamps correctly", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		sess := session.NewSession("timestamp-test")
		ctx := context.Background()

		// Create session first
		err := repo.Create(ctx, sess)
		require.NoError(t, err)

		// Get the original session from database to compare timestamps
		originalSess, err := repo.GetByID(ctx, sess.ID())
		require.NoError(t, err)
		originalUpdatedAt := originalSess.UpdatedAt()

		// Wait a bit and update
		time.Sleep(10 * time.Millisecond)
		err = sess.UpdateName("updated-name")
		require.NoError(t, err)

		// Act
		err = repo.Update(ctx, sess)
		require.NoError(t, err)

		// Assert
		retrievedSess, err := repo.GetByID(ctx, sess.ID())
		require.NoError(t, err)
		assert.Equal(t, "updated-name", retrievedSess.Name())
		// The session entity should have a newer timestamp than the original
		assert.True(t, sess.UpdatedAt().After(originalUpdatedAt),
			"Session entity UpdatedAt should be newer. Original: %v, Current: %v",
			originalUpdatedAt, sess.UpdatedAt())
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	t.Run("should delete session successfully", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		sess := session.NewSession("delete-test")
		ctx := context.Background()

		// Create session first
		err := repo.Create(ctx, sess)
		require.NoError(t, err)

		// Act
		err = repo.Delete(ctx, sess.ID())

		// Assert
		assert.NoError(t, err)

		// Verify session was deleted
		_, err = repo.GetByID(ctx, sess.ID())
		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionNotFound, err)
	})

	t.Run("should not fail when deleting non-existent session", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		nonExistentID := session.NewSessionID()
		ctx := context.Background()

		// Act
		err := repo.Delete(ctx, nonExistentID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionNotFound, err)
	})
}

func TestSessionRepository_List(t *testing.T) {
	t.Run("should list sessions with pagination", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Create multiple sessions
		sessions := make([]*session.Session, 5)
		for i := 0; i < 5; i++ {
			sess := session.NewSession(fmt.Sprintf("session-%d", i))
			sessions[i] = sess
			err := repo.Create(ctx, sess)
			require.NoError(t, err)
		}

		// Act
		retrievedSessions, total, err := repo.List(ctx, 3, 0)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, retrievedSessions, 3)
		assert.Equal(t, 5, total)

		// Verify sessions are ordered by created_at DESC
		for i := 0; i < len(retrievedSessions)-1; i++ {
			assert.True(t, retrievedSessions[i].CreatedAt().After(retrievedSessions[i+1].CreatedAt()) ||
				retrievedSessions[i].CreatedAt().Equal(retrievedSessions[i+1].CreatedAt()))
		}
	})

	t.Run("should handle pagination offset", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Create sessions
		for i := 0; i < 5; i++ {
			sess := session.NewSession(fmt.Sprintf("offset-session-%d", i))
			err := repo.Create(ctx, sess)
			require.NoError(t, err)
		}

		// Act - Get second page
		retrievedSessions, total, err := repo.List(ctx, 2, 2)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, retrievedSessions, 2)
		assert.Equal(t, 5, total)
	})

	t.Run("should return empty list when no sessions exist", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Act
		retrievedSessions, total, err := repo.List(ctx, 10, 0)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, retrievedSessions)
		assert.Equal(t, 0, total)
	})

	t.Run("should handle large offset", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Create one session
		sess := session.NewSession("single-session")
		err := repo.Create(ctx, sess)
		require.NoError(t, err)

		// Act - Request with offset larger than total
		retrievedSessions, total, err := repo.List(ctx, 10, 100)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, retrievedSessions)
		assert.Equal(t, 1, total)
	})
}

func TestSessionRepository_GetByStatus(t *testing.T) {
	t.Run("should get sessions by status", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Create sessions with different statuses
		disconnectedSess := session.NewSession("disconnected-session")
		connectedSess := session.NewSession("connected-session")
		err := connectedSess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		err = repo.Create(ctx, disconnectedSess)
		require.NoError(t, err)
		err = repo.Create(ctx, connectedSess)
		require.NoError(t, err)

		// Act - Get connected sessions
		connectedSessions, total, err := repo.GetByStatus(ctx, session.StatusConnected, 10, 0)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, connectedSessions, 1)
		assert.Equal(t, 1, total)
		assert.Equal(t, session.StatusConnected, connectedSessions[0].Status())
		assert.Equal(t, "connected-session", connectedSessions[0].Name())
	})

	t.Run("should return empty when no sessions match status", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Create only disconnected sessions
		sess := session.NewSession("disconnected-only")
		err := repo.Create(ctx, sess)
		require.NoError(t, err)

		// Act - Look for connected sessions
		connectedSessions, total, err := repo.GetByStatus(ctx, session.StatusConnected, 10, 0)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, connectedSessions)
		assert.Equal(t, 0, total)
	})
}

func TestSessionRepository_GetActiveCount(t *testing.T) {
	t.Run("should count active sessions correctly", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Create sessions - some active, some inactive
		activeSess1 := session.NewSession("active-1")
		err := activeSess1.Connect("active1@s.whatsapp.net")
		require.NoError(t, err)

		activeSess2 := session.NewSession("active-2")
		err = activeSess2.Connect("active2@s.whatsapp.net")
		require.NoError(t, err)

		inactiveSess := session.NewSession("inactive")

		err = repo.Create(ctx, activeSess1)
		require.NoError(t, err)
		err = repo.Create(ctx, activeSess2)
		require.NoError(t, err)
		err = repo.Create(ctx, inactiveSess)
		require.NoError(t, err)

		// Act
		activeCount, err := repo.GetActiveCount(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, activeCount)
	})

	t.Run("should return zero when no active sessions", func(t *testing.T) {
		// Arrange
		db := setupTestDB(t)
		defer db.Close()

		nullLogger := &NullLogger{}
		repo := repository.NewSessionRepository(db, nullLogger)
		ctx := context.Background()

		// Create only inactive sessions
		inactiveSess := session.NewSession("inactive-only")
		err := repo.Create(ctx, inactiveSess)
		require.NoError(t, err)

		// Act
		activeCount, err := repo.GetActiveCount(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, activeCount)
	})
}

// TestSessionRepository_MultiDatabase tests the repository with multiple databases
func TestSessionRepository_MultiDatabase(t *testing.T) {
	databases := setupTestDatabases(t)

	for _, testDB := range databases {
		t.Run(fmt.Sprintf("Database_%s", testDB.Name), func(t *testing.T) {
			// Setup
			setupSchema(t, testDB.DB)
			defer testDB.DB.Close()

			nullLogger := &NullLogger{}
			repo := repository.NewSessionRepository(testDB.DB, nullLogger)
			ctx := context.Background()

			t.Run("Create_and_Retrieve", func(t *testing.T) {
				sess := session.NewSession("multi-db-test")

				err := repo.Create(ctx, sess)
				assert.NoError(t, err)

				// Verify session was created
				retrieved, err := repo.GetByID(ctx, sess.ID())
				assert.NoError(t, err)
				assert.Equal(t, sess.ID(), retrieved.ID())
				assert.Equal(t, sess.Name(), retrieved.Name())
			})

			t.Run("CRUD_Operations", func(t *testing.T) {
				// Create
				sess := session.NewSession("crud-test")
				err := repo.Create(ctx, sess)
				require.NoError(t, err)

				// Read
				retrieved, err := repo.GetByName(ctx, "crud-test")
				require.NoError(t, err)
				assert.Equal(t, sess.ID(), retrieved.ID())

				// Update
				err = sess.Connect("test@s.whatsapp.net")
				require.NoError(t, err)
				err = repo.Update(ctx, sess)
				assert.NoError(t, err)

				// Verify update
				updated, err := repo.GetByID(ctx, sess.ID())
				assert.NoError(t, err)
				assert.Equal(t, session.StatusConnected, updated.Status())

				// Delete
				err = repo.Delete(ctx, sess.ID())
				assert.NoError(t, err)

				// Verify deletion
				_, err = repo.GetByID(ctx, sess.ID())
				assert.Equal(t, session.ErrSessionNotFound, err)
			})
		})
	}
}
