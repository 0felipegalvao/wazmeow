package database

import (
	"context"
	"fmt"
	"time"

	"wazmeow/internal/app/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// Database wraps the database connection and provides additional functionality
type Database struct {
	*sqlx.DB
}

// New creates a new database connection
func New(cfg config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

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
	// TODO: Implement migrations
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
