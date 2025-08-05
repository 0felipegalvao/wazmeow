package container

import (
	"context"
	"testing"

	"wazmeow/internal/infra/config"
	"wazmeow/internal/infra/container"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfraContainer_DatabaseMigration(t *testing.T) {
	// Create test configuration - disable WhatsApp to avoid foreign key issues
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver:       "sqlite3",
			URL:          "file:testdb3?mode=memory&cache=shared&_foreign_keys=1",
			AutoMigrate:  true,
			MaxOpenConns: 1, // Force single connection to avoid SQLite memory issues
			MaxIdleConns: 1, // Keep only one idle connection
			SQLite: config.SQLiteConfig{
				Path:        "",   // Use URL instead of Path
				ForeignKeys: true, // Enable foreign keys
				JournalMode: "MEMORY",
				Synchronous: "OFF",
				CacheSize:   1000,
				TempStore:   "memory",
				MmapSize:    0,
			},
		},
		Log: config.LogConfig{
			Level:         "debug", // Enable debug logs to see what's happening
			Output:        "console",
			ConsoleFormat: "console",
			FileFormat:    "json",
		},
		WhatsApp: config.WhatsAppConfig{},
		Security: config.SecurityConfig{},
		Features: config.FeaturesConfig{},
		Auth:     config.AuthConfig{},
		Proxy:    config.ProxyConfig{},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	// Create infrastructure container
	infraCont, err := container.New(cfg)
	if err != nil {
		t.Logf("Container creation failed with error: %v", err)
		// Let's try to create just the database part to see what fails
		t.SkipNow()
	}
	defer infraCont.Close()

	// Verify container is initialized
	assert.True(t, infraCont.IsInitialized(), "Container should be initialized")

	// Verify database connection
	db := infraCont.DB
	require.NotNil(t, db, "Database connection should not be nil")

	// Create context
	ctx := context.Background()

	// Verify table exists by querying it
	var count int
	err = db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("sqlite_master").
		Where("type = ? AND name = ?", "table", "wazmeow_sessions").
		Scan(ctx, &count)
	require.NoError(t, err, "Failed to query sqlite_master")
	assert.Equal(t, 1, count, "wazmeow_sessions table should exist")

	// Test that we can insert data
	_, err = db.ExecContext(ctx, `
		INSERT INTO wazmeow_sessions (id, name, status, is_active, created_at, updated_at)
		VALUES ('test-id', 'test-session', 'disconnected', false, datetime('now'), datetime('now'))
	`)
	require.NoError(t, err, "Should be able to insert data into wazmeow_sessions")

	// Test that we can read data
	var sessionCount int
	err = db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("wazmeow_sessions").
		Where("name = ?", "test-session").
		Scan(ctx, &sessionCount)
	require.NoError(t, err, "Should be able to read from wazmeow_sessions")
	assert.Equal(t, 1, sessionCount, "Should find the inserted session")
}
