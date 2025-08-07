package app

import (
	"context"
	"fmt"

	"wazmeow/internal/app/config"
	"wazmeow/internal/domain"
	"wazmeow/internal/services"
	"wazmeow/internal/storage"
	"wazmeow/internal/storage/repository"

	"github.com/rs/zerolog/log"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Container holds all application dependencies
type Container struct {
	config *config.Config
	db     *storage.Database

	// WhatsApp
	whatsappStoreManager *services.WhatsAppStoreManager
	multiSessionManager  *services.MultiSessionManager

	// Repositories
	sessionRepo domain.Repository

	// Use Cases
	createSessionUC *services.CreateSessionUseCase
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{
		config: cfg,
	}

	if err := container.initializeDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := container.initializeWhatsApp(); err != nil {
		return nil, fmt.Errorf("failed to initialize WhatsApp: %w", err)
	}

	if err := container.initializeRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := container.initializeMultiSessionManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize multi-session manager: %w", err)
	}

	if err := container.initializeUseCases(); err != nil {
		return nil, fmt.Errorf("failed to initialize use cases: %w", err)
	}

	log.Info().Msg("Application container initialized successfully")
	return container, nil
}

// initializeDatabase sets up the database connection and runs migrations
func (c *Container) initializeDatabase() error {
	db, err := storage.New(c.config.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	c.db = db
	log.Info().Msg("Database initialized successfully")
	return nil
}

// initializeWhatsApp sets up WhatsApp store and client managers
func (c *Container) initializeWhatsApp() error {
	// Create WhatsApp logger
	waLogger := waLog.Stdout("WhatsApp", "INFO", true)

	// Create WhatsApp store manager
	storeManager, err := services.NewWhatsAppStoreManager(c.db.DB.DB, waLogger)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp store manager: %w", err)
	}

	c.whatsappStoreManager = storeManager

	log.Info().Msg("WhatsApp initialized successfully")
	return nil
}

// initializeRepositories sets up all repositories
func (c *Container) initializeRepositories() error {
	c.sessionRepo = repository.NewSessionRepository(c.db.DB)

	log.Info().Msg("Repositories initialized successfully")
	return nil
}

// initializeMultiSessionManager sets up the multi-session manager
func (c *Container) initializeMultiSessionManager() error {
	// Create multi-session manager
	multiSessionManager := services.NewMultiSessionManager(c.whatsappStoreManager, c.sessionRepo)
	c.multiSessionManager = multiSessionManager

	log.Info().Msg("Multi-session manager initialized successfully")
	return nil
}

// initializeUseCases sets up all use cases
func (c *Container) initializeUseCases() error {
	c.createSessionUC = services.NewCreateSessionUseCase(c.sessionRepo)

	log.Info().Msg("Use cases initialized successfully")
	return nil
}

// Close closes all resources
func (c *Container) Close() error {
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
			return err
		}
	}

	log.Info().Msg("Application container closed successfully")
	return nil
}

// Getters for dependencies

func (c *Container) Config() *config.Config {
	return c.config
}

func (c *Container) Database() *storage.Database {
	return c.db
}

func (c *Container) SessionRepository() domain.Repository {
	return c.sessionRepo
}

func (c *Container) CreateSessionUseCase() *services.CreateSessionUseCase {
	return c.createSessionUC
}

func (c *Container) MultiSessionManager() *services.MultiSessionManager {
	return c.multiSessionManager
}

// func (c *Container) ConnectSessionUseCase() *sessionuc.ConnectSessionUseCase {
// 	return c.connectSessionUC
// }
