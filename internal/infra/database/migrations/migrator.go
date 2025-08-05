package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/bun"

	"wazmeow/internal/infra/database"
	"wazmeow/pkg/logger"
)

// Migrator handles database migrations
type Migrator struct {
	db     *bun.DB
	logger logger.Logger
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *bun.DB, log logger.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: log,
	}
}

// Migrate runs all database migrations
func (m *Migrator) Migrate(ctx context.Context) error {
	m.logger.Info("starting database migrations")

	// Create only our application table - whatsmeow will create its own tables
	models := []interface{}{
		(*database.WazMeowSessionModel)(nil),
	}

	for _, model := range models {
		if err := m.createTable(ctx, model); err != nil {
			return fmt.Errorf("failed to create table for model %T: %w", model, err)
		}
	}

	// Create indexes
	if err := m.createIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Create triggers for updated_at
	if err := m.createTriggers(ctx); err != nil {
		return fmt.Errorf("failed to create triggers: %w", err)
	}

	// Run schema migrations
	if err := m.runSchemaMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run schema migrations: %w", err)
	}

	m.logger.Info("database migrations completed successfully")
	return nil
}

// createTable creates a table if it doesn't exist
func (m *Migrator) createTable(ctx context.Context, model interface{}) error {
	// Log table creation with simple name extraction
	var tableName string
	switch model.(type) {
	case *database.WazMeowSessionModel:
		tableName = "wazmeow_sessions"
	default:
		tableName = "unknown"
	}

	m.logger.InfoWithFields("creating table", logger.Fields{
		"table": tableName,
	})

	// Use Bun's CreateTable
	query := m.db.NewCreateTable().
		Model(model).
		IfNotExists()

	// Log the SQL query for debugging
	sqlQuery, args := query.AppendQuery(m.db.Formatter(), nil)
	m.logger.DebugWithFields("executing create table query", logger.Fields{
		"table": tableName,
		"sql":   string(sqlQuery),
		"args":  args,
	})

	_, err := query.Exec(ctx)

	if err != nil {
		m.logger.ErrorWithError("failed to create table", err, logger.Fields{
			"table": tableName,
			"sql":   string(sqlQuery),
		})
		return err
	}

	// Table creation completed successfully
	m.logger.DebugWithFields("table creation completed", logger.Fields{
		"table": tableName,
	})

	m.logger.InfoWithFields("table created or verified", logger.Fields{
		"table": tableName,
	})

	return nil
}

// createIndexes creates database indexes
func (m *Migrator) createIndexes(ctx context.Context) error {
	indexes := []string{
		// WazMeow sessions table indexes
		"CREATE INDEX IF NOT EXISTS idx_wazmeow_sessions_name ON wazmeow_sessions(name)",
		"CREATE INDEX IF NOT EXISTS idx_wazmeow_sessions_status ON wazmeow_sessions(status)",
		"CREATE INDEX IF NOT EXISTS idx_wazmeow_sessions_is_active ON wazmeow_sessions(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_wazmeow_sessions_created_at ON wazmeow_sessions(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_wazmeow_sessions_wa_jid ON wazmeow_sessions(wa_jid)",
	}

	for _, indexSQL := range indexes {
		if _, err := m.db.ExecContext(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %s: %w", indexSQL, err)
		}
	}

	m.logger.InfoWithFields("database indexes created", logger.Fields{
		"count": len(indexes),
	})

	return nil
}

// createTriggers creates database triggers for automatic updated_at timestamps
func (m *Migrator) createTriggers(ctx context.Context) error {
	// Detect database type by checking dialect
	dialectName := fmt.Sprintf("%T", m.db.Dialect())

	var triggers []string

	switch dialectName {
	case "*sqlitedialect.Dialect":
		triggers = []string{
			// SQLite trigger for WazMeow sessions table
			`CREATE TRIGGER IF NOT EXISTS update_wazmeow_sessions_updated_at
			 AFTER UPDATE ON wazmeow_sessions
			 BEGIN
			   UPDATE wazmeow_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
			 END`,
		}
	case "*pgdialect.Dialect":
		// PostgreSQL uses functions and triggers differently
		triggers = []string{
			// Create function for updating timestamp
			`CREATE OR REPLACE FUNCTION update_updated_at_column()
			 RETURNS TRIGGER AS $$
			 BEGIN
			   NEW.updated_at = CURRENT_TIMESTAMP;
			   RETURN NEW;
			 END;
			 $$ language 'plpgsql'`,

			// Create trigger using the function
			`DROP TRIGGER IF EXISTS update_wazmeow_sessions_updated_at ON wazmeow_sessions`,
			`CREATE TRIGGER update_wazmeow_sessions_updated_at
			 BEFORE UPDATE ON wazmeow_sessions
			 FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()`,
		}
	default:
		m.logger.WarnWithFields("unknown database type, skipping triggers", logger.Fields{
			"database": dialectName,
		})
		return nil
	}

	for _, triggerSQL := range triggers {
		if _, err := m.db.ExecContext(ctx, triggerSQL); err != nil {
			return fmt.Errorf("failed to create trigger: %s: %w", triggerSQL, err)
		}
	}

	m.logger.InfoWithFields("database triggers created", logger.Fields{
		"count":    len(triggers),
		"database": dialectName,
	})

	return nil
}

// runSchemaMigrations runs schema migrations for adding new columns
func (m *Migrator) runSchemaMigrations(ctx context.Context) error {
	m.logger.Info("running schema migrations")

	// Detect database type by checking dialect
	dialectName := fmt.Sprintf("%T", m.db.Dialect())

	var migrations []string

	switch dialectName {
	case "*sqlitedialect.Dialect":
		migrations = []string{
			// Add proxy_config column to wazmeow_sessions table
			`ALTER TABLE wazmeow_sessions ADD COLUMN proxy_config TEXT DEFAULT NULL`,
		}
	case "*pgdialect.Dialect":
		migrations = []string{
			// Add proxy_config column to wazmeow_sessions table
			`ALTER TABLE wazmeow_sessions ADD COLUMN IF NOT EXISTS proxy_config JSONB DEFAULT NULL`,
		}
	default:
		m.logger.WarnWithFields("unknown database type, skipping schema migrations", logger.Fields{
			"database": dialectName,
		})
		return nil
	}

	for _, migrationSQL := range migrations {
		if _, err := m.db.ExecContext(ctx, migrationSQL); err != nil {
			// Check if error is about column already existing
			if strings.Contains(err.Error(), "duplicate column name") ||
				strings.Contains(err.Error(), "already exists") ||
				strings.Contains(err.Error(), "column already exists") {
				m.logger.InfoWithFields("column already exists, skipping migration", logger.Fields{
					"migration": migrationSQL,
				})
				continue
			}
			return fmt.Errorf("failed to run schema migration: %s: %w", migrationSQL, err)
		}
	}

	m.logger.InfoWithFields("schema migrations completed", logger.Fields{
		"count":    len(migrations),
		"database": dialectName,
	})

	return nil
}

// Drop drops all tables (useful for testing)
func (m *Migrator) Drop(ctx context.Context) error {
	m.logger.Warn("dropping all database tables")

	models := []interface{}{
		(*database.WazMeowSessionModel)(nil),
	}

	for _, model := range models {
		if err := m.dropTable(ctx, model); err != nil {
			return fmt.Errorf("failed to drop table for model %T: %w", model, err)
		}
	}

	m.logger.Info("all database tables dropped")
	return nil
}

// dropTable drops a table
func (m *Migrator) dropTable(ctx context.Context, model interface{}) error {
	_, err := m.db.NewDropTable().
		Model(model).
		IfExists().
		Exec(ctx)

	if err != nil {
		return err
	}

	// Log table drop with simple name extraction
	var tableName string
	switch model.(type) {
	case *database.WazMeowSessionModel:
		tableName = "wazmeow_sessions"
	default:
		tableName = "unknown"
	}

	m.logger.InfoWithFields("table dropped", logger.Fields{
		"table": tableName,
	})

	return nil
}

// Reset drops and recreates all tables
func (m *Migrator) Reset(ctx context.Context) error {
	m.logger.Warn("resetting database (drop and recreate all tables)")

	if err := m.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	if err := m.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to recreate tables: %w", err)
	}

	m.logger.Info("database reset completed")
	return nil
}
