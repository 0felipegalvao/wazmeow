package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"wazmeow/internal/infra/config"
	"wazmeow/pkg/logger"
)

// PostgreSQLConnection represents a PostgreSQL database connection
type PostgreSQLConnection struct {
	DB     *bun.DB
	Config *config.DatabaseConfig
	Logger logger.Logger
}

// NewPostgreSQLConnection creates a new PostgreSQL database connection
func NewPostgreSQLConnection(cfg *config.DatabaseConfig, log logger.Logger) (Connection, error) {
	conn := &PostgreSQLConnection{
		Config: cfg,
		Logger: log,
	}

	if err := conn.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return conn, nil
}

// GetDB returns the Bun DB instance
func (c *PostgreSQLConnection) GetDB() *bun.DB {
	return c.DB
}

// Close closes the database connection
func (c *PostgreSQLConnection) Close() error {
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			c.Logger.ErrorWithError("failed to close database connection", err, nil)
			return err
		}
		c.Logger.Info("database connection closed")
	}
	return nil
}

// Health checks the database health
func (c *PostgreSQLConnection) Health() error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := c.DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (c *PostgreSQLConnection) Stats() sql.DBStats {
	if c.DB == nil {
		return sql.DBStats{}
	}
	return c.DB.DB.Stats()
}

// connect establishes the PostgreSQL database connection
func (c *PostgreSQLConnection) connect() error {
	// Build connection string
	connStr, err := c.buildConnectionString()
	if err != nil {
		return fmt.Errorf("failed to build connection string: %w", err)
	}

	c.Logger.InfoWithFields("connecting to PostgreSQL", logger.Fields{
		"host":     c.Config.PostgreSQL.Host,
		"port":     c.Config.PostgreSQL.Port,
		"database": c.Config.PostgreSQL.Database,
		"username": c.Config.PostgreSQL.Username,
		"ssl_mode": c.Config.PostgreSQL.SSLMode,
	})

	// Open PostgreSQL connection
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(c.Config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.Config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(c.Config.ConnMaxLifetime)

	// Create Bun DB instance with PostgreSQL dialect
	c.DB = bun.NewDB(sqlDB, pgdialect.New())

	// Test the connection
	if err := c.DB.Ping(); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	// Set up PostgreSQL specific configurations
	if err := c.configurePostgreSQL(); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to configure PostgreSQL: %w", err)
	}

	c.Logger.InfoWithFields("PostgreSQL connection established", logger.Fields{
		"driver":            "postgres",
		"host":              c.Config.PostgreSQL.Host,
		"port":              c.Config.PostgreSQL.Port,
		"database":          c.Config.PostgreSQL.Database,
		"max_open_conns":    c.Config.MaxOpenConns,
		"max_idle_conns":    c.Config.MaxIdleConns,
		"conn_max_lifetime": c.Config.ConnMaxLifetime,
	})

	return nil
}

// buildConnectionString builds a PostgreSQL connection string
func (c *PostgreSQLConnection) buildConnectionString() (string, error) {
	// If URL is provided, use it directly
	if c.Config.URL != "" && c.Config.URL != "./data/wazmeow.db" {
		return c.Config.URL, nil
	}

	// Build connection string from individual components
	pgConfig := c.Config.PostgreSQL

	// Validate required fields
	if pgConfig.Host == "" {
		return "", fmt.Errorf("PostgreSQL host is required")
	}
	if pgConfig.Database == "" {
		return "", fmt.Errorf("PostgreSQL database name is required")
	}
	if pgConfig.Username == "" {
		return "", fmt.Errorf("PostgreSQL username is required")
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s",
		pgConfig.Host, pgConfig.Port, pgConfig.Database, pgConfig.Username)

	if pgConfig.Password != "" {
		connStr += fmt.Sprintf(" password=%s", pgConfig.Password)
	}

	if pgConfig.SSLMode != "" {
		connStr += fmt.Sprintf(" sslmode=%s", pgConfig.SSLMode)
	}

	if pgConfig.TimeZone != "" {
		connStr += fmt.Sprintf(" timezone=%s", url.QueryEscape(pgConfig.TimeZone))
	}

	if pgConfig.SearchPath != "" {
		connStr += fmt.Sprintf(" search_path=%s", pgConfig.SearchPath)
	}

	if pgConfig.AppName != "" {
		connStr += fmt.Sprintf(" application_name=%s", url.QueryEscape(pgConfig.AppName))
	}

	return connStr, nil
}

// configurePostgreSQL sets up PostgreSQL specific configurations
func (c *PostgreSQLConnection) configurePostgreSQL() error {
	ctx := context.Background()

	// Set timezone if specified
	if c.Config.PostgreSQL.TimeZone != "" {
		_, err := c.DB.ExecContext(ctx, "SET timezone = ?", c.Config.PostgreSQL.TimeZone)
		if err != nil {
			c.Logger.WarnWithError("failed to set timezone", err, logger.Fields{
				"timezone": c.Config.PostgreSQL.TimeZone,
			})
		}
	}

	// Set search path if specified
	if c.Config.PostgreSQL.SearchPath != "" {
		_, err := c.DB.ExecContext(ctx, "SET search_path = ?", c.Config.PostgreSQL.SearchPath)
		if err != nil {
			c.Logger.WarnWithError("failed to set search path", err, logger.Fields{
				"search_path": c.Config.PostgreSQL.SearchPath,
			})
		}
	}

	c.Logger.Info("PostgreSQL configuration applied successfully")
	return nil
}
