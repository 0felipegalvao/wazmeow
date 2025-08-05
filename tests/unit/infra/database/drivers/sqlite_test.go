package drivers_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/infra/config"
	"wazmeow/internal/infra/database/drivers"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteDriver_NewSQLiteConnection(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
	}{
		{
			name: "should create connection with memory database",
			config: &config.DatabaseConfig{
				Driver:          "sqlite3",
				URL:             ":memory:",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				SQLite: config.SQLiteConfig{
					ForeignKeys: true,
					JournalMode: "WAL",
					Synchronous: "NORMAL",
					CacheSize:   1000,
					TempStore:   "memory",
					MmapSize:    268435456,
				},
			},
			wantErr: false,
		},
		{
			name: "should create connection with file database",
			config: &config.DatabaseConfig{
				Driver:          "sqlite3",
				URL:             "./test_db.sqlite",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				SQLite: config.SQLiteConfig{
					Path:        "./test_db.sqlite",
					ForeignKeys: true,
					JournalMode: "WAL",
					Synchronous: "NORMAL",
					CacheSize:   1000,
					TempStore:   "memory",
					MmapSize:    268435456,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup before test
			if tt.config.SQLite.Path != "" && tt.config.SQLite.Path != ":memory:" {
				os.Remove(tt.config.SQLite.Path)
				defer os.Remove(tt.config.SQLite.Path)
			}

			conn, err := drivers.NewSQLiteConnection(tt.config, nil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, conn)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)

				// Test basic operations
				db := conn.GetDB()
				assert.NotNil(t, db)

				// Test health check
				err = conn.Health()
				assert.NoError(t, err)

				// Test stats
				stats := conn.Stats()
				assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)

				// Cleanup
				err = conn.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestSQLiteDriver_Connect_InvalidPath(t *testing.T) {
	tempDir := "./temp_invalid_test"
	dbPath := filepath.Join(tempDir, "subdir", "test.db")

	config := &config.DatabaseConfig{
		Driver:          "sqlite3",
		URL:             dbPath,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		SQLite: config.SQLiteConfig{
			Path:        dbPath,
			ForeignKeys: true,
			JournalMode: "WAL",
			Synchronous: "NORMAL",
			CacheSize:   1000,
			TempStore:   "memory",
			MmapSize:    268435456,
		},
	}

	// Ensure directory doesn't exist
	os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	conn, err := drivers.NewSQLiteConnection(config, nil)

	// Should work as SQLite will create the directory
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// Cleanup
	if conn != nil {
		conn.Close()
	}
}

func TestSQLiteDriver_DirectoryCreation(t *testing.T) {
	tempDir := "./temp_test_dir"
	dbPath := filepath.Join(tempDir, "test.db")

	config := &config.DatabaseConfig{
		Driver:          "sqlite3",
		URL:             dbPath,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		SQLite: config.SQLiteConfig{
			Path:        dbPath,
			ForeignKeys: true,
			JournalMode: "WAL",
			Synchronous: "NORMAL",
			CacheSize:   1000,
			TempStore:   "memory",
			MmapSize:    268435456,
		},
	}

	// Ensure directory doesn't exist
	os.RemoveAll(tempDir)
	defer os.RemoveAll(tempDir)

	conn, err := drivers.NewSQLiteConnection(config, nil)

	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// Check that directory was created
	_, err = os.Stat(tempDir)
	assert.NoError(t, err)

	// Cleanup
	conn.Close()
}

func TestSQLiteDriver_PragmaConfiguration(t *testing.T) {
	config := &config.DatabaseConfig{
		Driver:          "sqlite3",
		URL:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		SQLite: config.SQLiteConfig{
			ForeignKeys: true,
			JournalMode: "WAL",
			Synchronous: "NORMAL",
			CacheSize:   2000,
			TempStore:   "memory",
			MmapSize:    134217728, // 128MB
		},
	}

	conn, err := drivers.NewSQLiteConnection(config, nil)

	require.NoError(t, err)
	require.NotNil(t, conn)
	defer conn.Close()

	db := conn.GetDB()

	// Test that PRAGMA settings were applied
	var foreignKeys int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	assert.NoError(t, err)
	assert.Equal(t, 1, foreignKeys) // 1 means ON

	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	assert.NoError(t, err)
	// In-memory databases may not support WAL mode, so we check if it's either WAL or memory
	assert.Contains(t, []string{"wal", "memory"}, journalMode)

	var synchronous string
	err = db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	assert.NoError(t, err)
	assert.Equal(t, "1", synchronous) // NORMAL = 1

	var cacheSize int
	err = db.QueryRow("PRAGMA cache_size").Scan(&cacheSize)
	assert.NoError(t, err)
	assert.Equal(t, 2000, cacheSize)
}
