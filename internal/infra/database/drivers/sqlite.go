package drivers

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"

	"wazmeow/internal/infra/config"
	"wazmeow/pkg/logger"
)

// Connection interface (local copy to avoid import cycle)
type Connection interface {
	GetDB() *bun.DB
	Close() error
	Health() error
	Stats() sql.DBStats
}

// SQLiteConnection represents a SQLite database connection
type SQLiteConnection struct {
	DB     *bun.DB
	Config *config.DatabaseConfig
	Logger logger.Logger
}

// NewSQLiteConnection creates a new SQLite database connection
func NewSQLiteConnection(cfg *config.DatabaseConfig, log logger.Logger) (Connection, error) {
	conn := &SQLiteConnection{
		Config: cfg,
		Logger: log,
	}

	if err := conn.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	return conn, nil
}

// GetDB returns the Bun DB instance
func (c *SQLiteConnection) GetDB() *bun.DB {
	return c.DB
}

// Close closes the database connection
func (c *SQLiteConnection) Close() error {
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
func (c *SQLiteConnection) Health() error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := c.DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (c *SQLiteConnection) Stats() sql.DBStats {
	if c.DB == nil {
		return sql.DBStats{}
	}
	return c.DB.DB.Stats()
}

// connect establishes the SQLite database connection
func (c *SQLiteConnection) connect() error {
	dbPath := c.Config.URL
	if c.Config.SQLite.Path != "" {
		dbPath = c.Config.SQLite.Path
	}

	// Ensure the directory exists (only for file-based databases, not in-memory)
	if !strings.Contains(dbPath, ":memory:") && !strings.Contains(dbPath, "mode=memory") {
		dir := filepath.Dir(dbPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}
	}

	// Open SQLite connection
	sqlDB, err := sql.Open(sqliteshim.ShimName, dbPath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(c.Config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.Config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(c.Config.ConnMaxLifetime)

	// Create Bun DB instance with SQLite dialect
	c.DB = bun.NewDB(sqlDB, sqlitedialect.New())

	// Test the connection
	if err := c.DB.Ping(); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Apply SQLite specific configurations
	if err := c.configureSQLite(sqlDB); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to configure SQLite: %w", err)
	}

	c.Logger.InfoWithFields("SQLite connection established", logger.Fields{
		"driver":            c.Config.Driver,
		"path":              dbPath,
		"max_open_conns":    c.Config.MaxOpenConns,
		"max_idle_conns":    c.Config.MaxIdleConns,
		"conn_max_lifetime": c.Config.ConnMaxLifetime,
	})

	return nil
}

// configureSQLite applies SQLite specific configurations
func (c *SQLiteConnection) configureSQLite(sqlDB *sql.DB) error {
	sqliteConfig := c.Config.SQLite

	// Build PRAGMA statements based on configuration
	pragmas := []string{}

	// Foreign keys
	if sqliteConfig.ForeignKeys {
		pragmas = append(pragmas, "PRAGMA foreign_keys = ON")
	} else {
		pragmas = append(pragmas, "PRAGMA foreign_keys = OFF")
	}

	// Journal mode
	journalMode := sqliteConfig.JournalMode
	if journalMode == "" {
		journalMode = "WAL"
	}
	pragmas = append(pragmas, fmt.Sprintf("PRAGMA journal_mode = %s", journalMode))

	// Synchronous mode
	synchronous := sqliteConfig.Synchronous
	if synchronous == "" {
		synchronous = "NORMAL"
	}
	pragmas = append(pragmas, fmt.Sprintf("PRAGMA synchronous = %s", synchronous))

	// Cache size
	cacheSize := sqliteConfig.CacheSize
	if cacheSize == 0 {
		cacheSize = 1000
	}
	pragmas = append(pragmas, fmt.Sprintf("PRAGMA cache_size = %d", cacheSize))

	// Temp store
	tempStore := sqliteConfig.TempStore
	if tempStore == "" {
		tempStore = "memory"
	}
	pragmas = append(pragmas, fmt.Sprintf("PRAGMA temp_store = %s", tempStore))

	// Memory-mapped I/O
	mmapSize := sqliteConfig.MmapSize
	if mmapSize == 0 {
		mmapSize = 268435456 // 256MB
	}
	pragmas = append(pragmas, fmt.Sprintf("PRAGMA mmap_size = %d", mmapSize))

	// Execute all PRAGMA statements
	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	c.Logger.InfoWithFields("SQLite configuration applied", logger.Fields{
		"foreign_keys": sqliteConfig.ForeignKeys,
		"journal_mode": journalMode,
		"synchronous":  synchronous,
		"cache_size":   cacheSize,
		"temp_store":   tempStore,
		"mmap_size":    mmapSize,
	})

	return nil
}
