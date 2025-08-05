package database

import (
	"context"
	"database/sql"
	"testing"

	"wazmeow/internal/infra/database"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

func setupTestDB(t *testing.T) *bun.DB {
	// Create in-memory SQLite database for testing
	sqldb, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create bun DB instance
	db := bun.NewDB(sqldb, sqlitedialect.New())
	return db
}

func TestWazMeowSessionModel_CreateTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Test creating table using Bun's CreateTable
	_, err := db.NewCreateTable().
		Model((*database.WazMeowSessionModel)(nil)).
		IfNotExists().
		Exec(ctx)
	require.NoError(t, err, "Should be able to create wazmeow_sessions table")

	// Verify table exists
	var count int
	err = db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("sqlite_master").
		Where("type = ? AND name = ?", "table", "wazmeow_sessions").
		Scan(ctx, &count)
	require.NoError(t, err, "Failed to query sqlite_master")
	assert.Equal(t, 1, count, "wazmeow_sessions table should exist")

	// Verify table structure by getting column info
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(wazmeow_sessions)")
	require.NoError(t, err)
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		require.NoError(t, err)
		columns[name] = true
	}

	// Check required columns exist
	expectedColumns := []string{
		"id", "name", "status", "wa_jid", "qr_code",
		"proxy_config", "is_active", "created_at", "updated_at",
	}

	for _, col := range expectedColumns {
		assert.True(t, columns[col], "Column %s should exist", col)
	}

	// Test inserting data
	_, err = db.ExecContext(ctx, `
		INSERT INTO wazmeow_sessions (id, name, status, is_active, created_at, updated_at)
		VALUES ('test-id', 'test-session', 'disconnected', false, datetime('now'), datetime('now'))
	`)
	require.NoError(t, err, "Should be able to insert data")

	// Test reading data
	var sessionCount int
	err = db.NewSelect().
		ColumnExpr("COUNT(*)").
		TableExpr("wazmeow_sessions").
		Where("name = ?", "test-session").
		Scan(ctx, &sessionCount)
	require.NoError(t, err, "Should be able to read from wazmeow_sessions")
	assert.Equal(t, 1, sessionCount, "Should find the inserted session")
}
