package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wazmeow/internal/app/container"
	"wazmeow/internal/infra/config"
	sessionUC "wazmeow/internal/usecases/session"
	"wazmeow/pkg/logger"
)

// App represents the main application
type App struct {
	container *container.AppContainer
	logger    logger.Logger
}

// New creates a new application instance
func New(opts ...container.AppOption) (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create application container with options
	appContainer, err := container.NewAppContainer(cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create app container: %w", err)
	}

	return &App{
		container: appContainer,
		logger:    appContainer.GetLogger(),
	}, nil
}

// Start starts the application with graceful shutdown
func (a *App) Start() error {
	a.logger.Info("Starting WazMeow application")

	// Start WhatsApp manager
	if err := a.startWhatsAppManager(); err != nil {
		return err
	}

	// Perform automatic reconnection of sessions
	a.performAutoReconnection()

	// Start server and wait for shutdown
	return a.startServerAndWaitForShutdown()
}

// Stop stops the application
func (a *App) Stop() error {
	a.logger.Info("Stopping WazMeow application")

	if err := a.container.Close(); err != nil {
		return fmt.Errorf("failed to close app container: %w", err)
	}

	a.logger.Info("WazMeow application stopped successfully")
	return nil
}

// Health checks the application health
func (a *App) Health() error {
	return a.container.Health()
}

// GetConfig returns the application configuration
func (a *App) GetConfig() *config.Config {
	return a.container.GetConfig()
}

// GetContainer returns the application container
func (a *App) GetContainer() *container.AppContainer {
	return a.container
}

// GetServerInfo returns information about the HTTP server
func (a *App) GetServerInfo() interface{} {
	return a.container.GetServerInfo()
}

// GetStats returns application statistics
func (a *App) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"server":   a.container.GetServerInfo(),
		"database": a.container.GetDatabaseStats(),
		"whatsapp": a.container.GetWhatsAppStats(),
	}
}

// startWhatsAppManager starts the WhatsApp manager
func (a *App) startWhatsAppManager() error {
	if err := a.container.StartWhatsAppManager(); err != nil {
		return fmt.Errorf("failed to start WhatsApp manager: %w", err)
	}
	return nil
}

// performAutoReconnection handles automatic session reconnection during startup
func (a *App) performAutoReconnection() {
	a.logger.Info("🔄 Starting automatic session reconnection")
	reconnectCtx, reconnectCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer reconnectCancel()

	reconnectResult, err := a.container.AutoReconnectSessions(reconnectCtx)
	if err != nil {
		a.logger.ErrorWithError("failed to perform automatic reconnection", err, nil)
		// Don't fail the startup, just log the error
		return
	}

	a.logReconnectionResults(reconnectResult)
}

// logReconnectionResults logs the results of automatic reconnection
func (a *App) logReconnectionResults(result interface{}) {
	// Import the session use case package to access the types
	if reconnectResult, ok := result.(*sessionUC.AutoReconnectResponse); ok {
		a.logger.InfoWithFields("automatic reconnection completed", logger.Fields{
			"total_sessions":           reconnectResult.TotalSessions,
			"eligible_sessions":        reconnectResult.EligibleSessions,
			"successful_reconnections": reconnectResult.SuccessfulReconnections,
			"failed_reconnections":     reconnectResult.FailedReconnections,
		})

		// Log individual session results
		for _, sessionResult := range reconnectResult.ReconnectionResults {
			a.logSessionReconnectionResult(sessionResult)
		}
	}
}

// logSessionReconnectionResult logs individual session reconnection results
func (a *App) logSessionReconnectionResult(result sessionUC.SessionReconnectionResult) {
	if result.Success {
		a.logger.InfoWithFields("✅ session reconnected", logger.Fields{
			"session_id":   result.SessionID.String(),
			"session_name": result.SessionName,
			"duration_ms":  result.Duration.Milliseconds(),
		})
	} else {
		a.logger.WarnWithFields("❌ session reconnection failed", logger.Fields{
			"session_id":   result.SessionID.String(),
			"session_name": result.SessionName,
			"error":        result.Error,
			"duration_ms":  result.Duration.Milliseconds(),
		})
	}
}

// startServerAndWaitForShutdown starts the HTTP server and waits for shutdown signals
func (a *App) startServerAndWaitForShutdown() error {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		if err := a.container.StartServer(ctx); err != nil {
			serverErrors <- err
		}
	}()

	a.logger.InfoWithFields("WazMeow application started successfully", logger.Fields{
		"server_address": a.container.GetServerInfo().Address,
	})

	// Wait for shutdown signal or server error
	return a.waitForShutdown(serverErrors, sigChan, cancel)
}

// waitForShutdown waits for either a server error or shutdown signal
func (a *App) waitForShutdown(serverErrors <-chan error, sigChan <-chan os.Signal, cancel context.CancelFunc) error {
	select {
	case err := <-serverErrors:
		a.logger.ErrorWithError("Server error", err, nil)
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		a.logger.InfoWithFields("Shutdown signal received", logger.Fields{
			"signal": sig.String(),
		})
		cancel() // Cancel context to trigger graceful shutdown
		return nil
	}
}
