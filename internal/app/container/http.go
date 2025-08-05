package container

import (
	"context"
	"fmt"

	"wazmeow/internal/http/handler"
	"wazmeow/internal/http/routes"
	"wazmeow/internal/http/server"
	"wazmeow/internal/infra/config"
	"wazmeow/internal/infra/container"
	"wazmeow/pkg/logger"
)

// httpContainer implements HTTPContainer interface
type httpContainer struct {
	sessionHandler *handler.SessionHandler
	healthHandler  *handler.HealthHandler
	router         *routes.Router
	httpServer     *server.Server
	serverManager  *server.ServerManager
	logger         logger.Logger
	isInitialized  bool
}

// NewHTTPContainer creates a new HTTP container
func NewHTTPContainer(
	infraContainer *container.Container,
	useCaseContainer UseCaseContainer,
	cfg *config.Config,
) (HTTPContainer, error) {
	hc := &httpContainer{
		logger: infraContainer.Logger,
	}

	if err := hc.initialize(infraContainer, useCaseContainer, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP container: %w", err)
	}

	return hc, nil
}

// initialize sets up HTTP layer components
func (hc *httpContainer) initialize(
	infraContainer *container.Container,
	useCaseContainer UseCaseContainer,
	cfg *config.Config,
) error {
	logger := infraContainer.Logger
	validator := infraContainer.Validator

	sessionUseCases := useCaseContainer.GetSessionUseCases()
	whatsappUseCases := useCaseContainer.GetWhatsAppUseCases()

	// Create handlers
	hc.sessionHandler = handler.NewSessionHandler(
		sessionUseCases.Create,
		sessionUseCases.Connect,
		sessionUseCases.Disconnect,
		sessionUseCases.List,
		sessionUseCases.Delete,
		sessionUseCases.Resolve,
		sessionUseCases.SetProxy,
		whatsappUseCases.GenerateQR,
		whatsappUseCases.PairPhone,
		logger,
		validator,
	)

	hc.healthHandler = handler.NewHealthHandler(
		infraContainer,
		logger,
	)

	// Create router
	hc.router = routes.NewRouter(
		hc.sessionHandler,
		hc.healthHandler,
		cfg,
		logger,
	)

	// Create HTTP server
	hc.httpServer = server.New(
		hc.router,
		&cfg.Server,
		logger,
	)

	// Create server manager
	hc.serverManager = server.NewServerManager(
		hc.httpServer,
		logger,
	)

	hc.isInitialized = true
	logger.Info("HTTP container initialized successfully")
	return nil
}

// GetServerManager returns the server manager
func (hc *httpContainer) GetServerManager() *server.ServerManager {
	return hc.serverManager
}

// GetServerInfo returns server information
func (hc *httpContainer) GetServerInfo() server.ServerInfo {
	if hc.serverManager != nil {
		return hc.serverManager.GetServerInfo()
	}
	return server.ServerInfo{}
}

// StartServer starts the HTTP server
func (hc *httpContainer) StartServer(ctx context.Context) error {
	if !hc.isInitialized {
		return fmt.Errorf("HTTP container not initialized")
	}

	hc.logger.InfoWithFields("Starting HTTP server", logger.Fields{
		"address": hc.httpServer.GetAddr(),
	})

	return hc.serverManager.StartWithGracefulShutdown(ctx)
}
