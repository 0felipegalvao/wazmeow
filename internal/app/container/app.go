package container

import (
	"context"
	"fmt"
	"time"

	"wazmeow/internal/http/server"
	"wazmeow/internal/infra/config"
	"wazmeow/internal/infra/container"
	"wazmeow/internal/usecases/session"
	"wazmeow/pkg/logger"
)

// AppContainer holds all application dependencies
type AppContainer struct {
	// Configuration
	config  *config.Config
	options *AppOptions

	// Infrastructure
	infraContainer *container.Container

	// Sub-containers
	useCaseContainer UseCaseContainer
	httpContainer    HTTPContainer

	// Internal state
	isInitialized bool
}

// NewAppContainer creates a new application container
func NewAppContainer(cfg *config.Config, opts ...AppOption) (*AppContainer, error) {
	// Apply default options
	options := DefaultAppOptions()

	// Apply custom options
	for _, opt := range opts {
		opt(options)
	}

	appContainer := &AppContainer{
		config:  cfg,
		options: options,
	}

	if err := appContainer.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize app container: %w", err)
	}

	return appContainer, nil
}

// initialize sets up all application dependencies
func (c *AppContainer) initialize() error {
	// Initialize infrastructure container
	if err := c.initializeInfrastructure(); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize use cases
	if err := c.initializeUseCases(); err != nil {
		return fmt.Errorf("failed to initialize use cases: %w", err)
	}

	// Initialize HTTP layer
	if err := c.initializeHTTPLayer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP layer: %w", err)
	}

	c.isInitialized = true
	c.infraContainer.Logger.Info("Application container initialized successfully")

	return nil
}

// initializeInfrastructure sets up infrastructure dependencies
func (c *AppContainer) initializeInfrastructure() error {
	infraContainer, err := container.New(c.config)
	if err != nil {
		return fmt.Errorf("failed to create infrastructure container: %w", err)
	}

	c.infraContainer = infraContainer
	return nil
}

// initializeUseCases sets up use case dependencies
func (c *AppContainer) initializeUseCases() error {
	useCaseContainer, err := NewUseCaseContainer(c.infraContainer)
	if err != nil {
		return fmt.Errorf("failed to create use case container: %w", err)
	}

	c.useCaseContainer = useCaseContainer
	return nil
}

// initializeHTTPLayer sets up HTTP layer dependencies
func (c *AppContainer) initializeHTTPLayer() error {
	httpContainer, err := NewHTTPContainer(c.infraContainer, c.useCaseContainer, c.config)
	if err != nil {
		return fmt.Errorf("failed to create HTTP container: %w", err)
	}

	c.httpContainer = httpContainer
	return nil
}

// GetLogger returns the application logger
func (c *AppContainer) GetLogger() logger.Logger {
	if c.infraContainer != nil {
		return c.infraContainer.Logger
	}
	return nil
}

// GetConfig returns the application configuration
func (c *AppContainer) GetConfig() *config.Config {
	return c.config
}

// GetServerManager returns the server manager
func (c *AppContainer) GetServerManager() *server.ServerManager {
	return c.httpContainer.GetServerManager()
}

// GetInfraContainer returns the infrastructure container
func (c *AppContainer) GetInfraContainer() *container.Container {
	return c.infraContainer
}

// IsInitialized returns true if the container is initialized
func (c *AppContainer) IsInitialized() bool {
	return c.isInitialized
}

// Close gracefully shuts down all dependencies
func (c *AppContainer) Close() error {
	if !c.isInitialized {
		return nil
	}

	c.infraContainer.Logger.Info("Shutting down application container")

	var errors []error

	// Close infrastructure container
	if c.infraContainer != nil {
		if err := c.infraContainer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close infrastructure container: %w", err))
		}
	}

	if len(errors) > 0 {
		// Log all errors
		for _, err := range errors {
			c.infraContainer.Logger.ErrorWithError("error during app container shutdown", err, nil)
		}
		return fmt.Errorf("multiple errors during shutdown: %v", errors)
	}

	c.infraContainer.Logger.Info("Application container shut down successfully")
	return nil
}

// Health checks the health of all application components
func (c *AppContainer) Health() error {
	if !c.isInitialized {
		return fmt.Errorf("container not initialized")
	}

	// Check infrastructure health
	if err := c.infraContainer.Health(); err != nil {
		return fmt.Errorf("infrastructure health check failed: %w", err)
	}

	return nil
}

// GetServerInfo returns information about the HTTP server
func (c *AppContainer) GetServerInfo() server.ServerInfo {
	return c.httpContainer.GetServerInfo()
}

// StartWhatsAppManager starts the WhatsApp manager
func (c *AppContainer) StartWhatsAppManager() error {
	return c.infraContainer.StartWhatsAppManager()
}

// GetDatabaseStats returns database statistics
func (c *AppContainer) GetDatabaseStats() any {
	return c.infraContainer.GetDatabaseStats()
}

// GetWhatsAppStats returns WhatsApp statistics
func (c *AppContainer) GetWhatsAppStats() any {
	return c.infraContainer.GetWhatsAppStats()
}

// AutoReconnectSessions performs automatic reconnection of eligible sessions
func (c *AppContainer) AutoReconnectSessions(ctx context.Context) (*session.AutoReconnectResponse, error) {
	if !c.isInitialized {
		return nil, fmt.Errorf("container not initialized")
	}

	sessionUseCases := c.useCaseContainer.GetSessionUseCases()

	req := session.AutoReconnectRequest{
		MaxConcurrentReconnections: 5, // Limit concurrent reconnections
		ReconnectionTimeout:        30 * time.Second,
	}

	return sessionUseCases.AutoReconnect.Execute(ctx, req)
}

// StartServer starts the HTTP server
func (c *AppContainer) StartServer(ctx context.Context) error {
	return c.httpContainer.StartServer(ctx)
}
