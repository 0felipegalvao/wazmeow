package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"wazmeow/internal/app/config"
	"wazmeow/internal/domain"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// Database wraps the database connection and provides additional functionality
type Database struct {
	*bun.DB
}

// New creates a new database connection
func New(cfg config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)

	// Create Bun database connection
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// Add debug hook if enabled
	if cfg.Debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// Configure connection pool
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		DB: db,
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Msg("Database connected successfully")

	return database, nil
}

// Migrate runs all pending database migrations
func (d *Database) Migrate(ctx context.Context) error {
	log.Info().Msg("Starting database migration")

	// Auto-create sessions table
	_, err := d.NewCreateTable().
		Model((*domain.Session)(nil)).
		IfNotExists().
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create sessions table")
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	log.Info().Msg("Database migration completed successfully")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	log.Info().Msg("Closing database connection")
	return d.DB.Close()
}

// Health checks the database health
func (d *Database) Health(ctx context.Context) error {
	return d.PingContext(ctx)
}
