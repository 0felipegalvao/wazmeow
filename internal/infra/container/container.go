package container

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver for whatsmeow
	"github.com/uptrace/bun"
	"go.mau.fi/whatsmeow/store/sqlstore"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/internal/infra/config"
	"wazmeow/internal/infra/database"
	"wazmeow/internal/infra/database/migrations"
	infraLogger "wazmeow/internal/infra/logger"
	"wazmeow/internal/infra/repository"
	"wazmeow/internal/infra/whats"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// Container holds all infrastructure dependencies
type Container struct {
	// Configuration
	Config *config.Config

	// Core infrastructure
	Logger    logger.Logger
	Validator validator.Validator
	DB        *bun.DB

	// Database components
	DBConnection database.Connection
	Migrator     *migrations.Migrator

	// Repositories
	SessionRepo session.Repository

	// WhatsApp components
	WhatsAppStore   *sqlstore.Container
	WhatsAppManager whatsapp.Manager

	// Internal state
	isInitialized bool
}

// New creates a new infrastructure container
func New(cfg *config.Config) (*Container, error) {
	container := &Container{
		Config: cfg,
	}

	if err := container.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	return container, nil
}

// initialize sets up all infrastructure components
func (c *Container) initialize() error {
	// Initialize logger first
	if err := c.initializeLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	c.Logger.Info("initializing infrastructure container")

	// Initialize validator
	if err := c.initializeValidator(); err != nil {
		return fmt.Errorf("failed to initialize validator: %w", err)
	}

	// Initialize database
	if err := c.initializeDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	if err := c.initializeRepositories(); err != nil {
		return fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Initialize WhatsApp manager
	if err := c.initializeWhatsApp(); err != nil {
		return fmt.Errorf("failed to initialize WhatsApp: %w", err)
	}

	c.isInitialized = true
	c.Logger.Info("infrastructure container initialized successfully")

	return nil
}

// initializeLogger sets up the logger
func (c *Container) initializeLogger() error {
	c.Logger = infraLogger.New(&c.Config.Log)
	return nil
}

// initializeValidator sets up the validator
func (c *Container) initializeValidator() error {
	c.Validator = validator.New()
	return nil
}

// initializeDatabase sets up the database connection and migrations
func (c *Container) initializeDatabase() error {
	// Create database connection
	dbConn, err := database.New(&c.Config.Database, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

	c.DBConnection = dbConn
	c.DB = dbConn.GetDB()

	// Create migrator
	c.Migrator = migrations.NewMigrator(c.DB, c.Logger)

	// Run migrations if auto-migrate is enabled
	if c.Config.Database.AutoMigrate {
		ctx := context.Background()
		if err := c.Migrator.Migrate(ctx); err != nil {
			return fmt.Errorf("failed to run database migrations: %w", err)
		}
	}

	return nil
}

// initializeRepositories sets up all repositories
func (c *Container) initializeRepositories() error {
	// Session repository
	c.SessionRepo = repository.NewSessionRepository(c.DB, c.Logger)

	c.Logger.Info("repositories initialized")
	return nil
}

// initializeWhatsApp sets up WhatsApp components
func (c *Container) initializeWhatsApp() error {
	// Create WhatsApp sqlstore container using the same database
	dbURL := c.Config.Database.URL
	dbDriver := c.Config.Database.Driver

	// Adjust driver name for whatsmeow compatibility
	switch dbDriver {
	case "sqlite", "sqlite3":
		dbDriver = "sqlite3"
		// Add foreign keys parameter for SQLite (only for file-based databases)
		if dbURL == "./data/wazmeow.db" {
			dbURL = "./data/wazmeow.db?_foreign_keys=on"
		} else if !strings.Contains(dbURL, ":memory:") && !strings.Contains(dbURL, "mode=memory") && !strings.Contains(dbURL, "_foreign_keys") {
			// Add foreign keys parameter if not already present and not in-memory
			if strings.Contains(dbURL, "?") {
				dbURL += "&_foreign_keys=on"
			} else {
				dbURL += "?_foreign_keys=on"
			}
		}
	case "postgres", "postgresql":
		dbDriver = "postgres"
	default:
		return fmt.Errorf("unsupported database driver for WhatsApp store: %s", dbDriver)
	}

	// Create logger adapter for whatsmeow
	waLogger := whats.NewLoggerAdapter(c.Logger, "WhatsApp")

	whatsappStore, err := sqlstore.New(context.Background(), dbDriver, dbURL, waLogger)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp store: %w", err)
	}

	// Upgrade WhatsApp store schema
	err = whatsappStore.Upgrade(context.Background())
	if err != nil {
		return fmt.Errorf("failed to upgrade WhatsApp store: %w", err)
	}

	c.WhatsAppStore = whatsappStore

	// Create WhatsApp manager
	c.WhatsAppManager = whats.NewManager(&c.Config.WhatsApp, whatsappStore, c.SessionRepo, c.Logger)

	c.Logger.Info("WhatsApp components initialized")
	return nil
}

// Close gracefully shuts down all infrastructure components
func (c *Container) Close() error {
	if !c.isInitialized {
		return nil
	}

	c.Logger.Info("shutting down infrastructure container")

	var errors []error

	// Stop WhatsApp manager
	if c.WhatsAppManager != nil {
		if err := c.WhatsAppManager.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop WhatsApp manager: %w", err))
		}
	}

	// Close WhatsApp store
	if c.WhatsAppStore != nil {
		if err := c.WhatsAppStore.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close WhatsApp store: %w", err))
		}
	}

	// Close database connection
	if c.DBConnection != nil {
		if err := c.DBConnection.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database connection: %w", err))
		}
	}

	if len(errors) > 0 {
		// Log all errors
		for _, err := range errors {
			c.Logger.ErrorWithError("error during container shutdown", err, nil)
		}
		return fmt.Errorf("multiple errors during shutdown: %v", errors)
	}

	c.Logger.Info("infrastructure container shut down successfully")
	return nil
}

// Health checks the health of all infrastructure components
func (c *Container) Health() error {
	if !c.isInitialized {
		return fmt.Errorf("container not initialized")
	}

	// Check database health
	if err := c.DBConnection.Health(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check WhatsApp manager health
	if err := c.WhatsAppManager.HealthCheck(); err != nil {
		return fmt.Errorf("WhatsApp manager health check failed: %w", err)
	}

	return nil
}

// IsInitialized returns true if the container is initialized
func (c *Container) IsInitialized() bool {
	return c.isInitialized
}

// GetDatabaseStats returns database connection statistics
func (c *Container) GetDatabaseStats() interface{} {
	if c.DB == nil {
		return sql.DBStats{}
	}
	return c.DB.DB.Stats()
}

// GetWhatsAppStats returns WhatsApp manager statistics
func (c *Container) GetWhatsAppStats() *whatsapp.ManagerStats {
	if c.WhatsAppManager == nil {
		return nil
	}
	// Cast to concrete type to access GetStats method
	if manager, ok := c.WhatsAppManager.(*whats.Manager); ok {
		return manager.GetStats()
	}
	return nil
}

// StartWhatsAppManager starts the WhatsApp manager
func (c *Container) StartWhatsAppManager() error {
	if c.WhatsAppManager == nil {
		return fmt.Errorf("WhatsApp manager not initialized")
	}

	ctx := context.Background()
	return c.WhatsAppManager.Start(ctx)
}

// ResetDatabase drops and recreates all database tables
func (c *Container) ResetDatabase() error {
	if c.Migrator == nil {
		return fmt.Errorf("migrator not initialized")
	}

	c.Logger.Warn("resetting database")
	ctx := context.Background()
	return c.Migrator.Reset(ctx)
}

// MigrateDatabase runs database migrations
func (c *Container) MigrateDatabase() error {
	if c.Migrator == nil {
		return fmt.Errorf("migrator not initialized")
	}

	c.Logger.Info("running database migrations")
	ctx := context.Background()
	return c.Migrator.Migrate(ctx)
}
