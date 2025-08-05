package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"

	"wazmeow/internal/infra/config"
	"wazmeow/internal/infra/database/drivers"
	"wazmeow/pkg/logger"
)

// Connection interface defines the database connection contract
type Connection interface {
	GetDB() *bun.DB
	Close() error
	Health() error
	Stats() sql.DBStats
}

// DriverConnection interface for driver implementations
type DriverConnection interface {
	GetDB() *bun.DB
	Close() error
	Health() error
	Stats() sql.DBStats
}

// Migrator interface defines database migration operations
type Migrator interface {
	Migrate(ctx context.Context) error
	Reset(ctx context.Context) error
	Drop(ctx context.Context) error
}

// DatabaseType represents the type of database
type DatabaseType string

const (
	SQLite      DatabaseType = "sqlite"
	SQLite3     DatabaseType = "sqlite3"
	PostgreSQL  DatabaseType = "postgres"
	PostgreSQL2 DatabaseType = "postgresql"
)

// ConnectionFactory creates database connections based on configuration
type ConnectionFactory struct {
	logger logger.Logger
}

// NewConnectionFactory creates a new connection factory
func NewConnectionFactory(log logger.Logger) *ConnectionFactory {
	return &ConnectionFactory{
		logger: log,
	}
}

// CreateConnection creates a database connection based on the driver type
func (f *ConnectionFactory) CreateConnection(cfg *config.DatabaseConfig) (Connection, error) {
	dbType := DatabaseType(cfg.Driver)

	if f.logger != nil {
		f.logger.InfoWithFields("creating database connection", logger.Fields{
			"driver": cfg.Driver,
			"type":   string(dbType),
		})
	}

	switch dbType {
	case SQLite, SQLite3:
		return drivers.NewSQLiteConnection(cfg, f.logger)
	case PostgreSQL, PostgreSQL2:
		return drivers.NewPostgreSQLConnection(cfg, f.logger)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// New creates a new database connection using the factory pattern
func New(cfg *config.DatabaseConfig, log logger.Logger) (Connection, error) {
	factory := NewConnectionFactory(log)
	return factory.CreateConnection(cfg)
}
